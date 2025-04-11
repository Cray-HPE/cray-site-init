/*
 * MIT License
 *
 * (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package pit

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SealedSecret - Minimum struct to determine secret scope
// and access encrypted data
type SealedSecret struct {
	Spec struct {
		EncryptedData map[string]string `yaml:"encryptedData"`
		Template      struct {
			Metadata struct {
				Annotations map[string]string
			}
		}
	}
}

// Customizations - Minimum customizations (shasta-cfg) struct
// to access sealed secrets
type Customizations struct {
	Spec struct {
		Kubernetes struct {
			SealedSecrets map[string]SealedSecret `yaml:"sealed_secrets"`
		}
	}
}

// CACerts - For storage of ca-certs cloud-init update
type CACerts struct {
	RemoveDefaults bool     `json:"remove-defaults"`
	Trusted        []string `json:"trusted"`
}

// Metadata - Boilerplate for cloud-init hierarchical structure
type Metadata struct {
	CACerts `json:"ca-certs"`
}

// Global - Boilerplate for cloud-init hierarchical structure
type Global struct {
	Global struct {
		Metadata `json:"meta-data"`
	}
}

var customizationsFile string
var sealedSecretsKeyFile string
var sealedSecretName string

func caCommand() *cobra.Command {
	c := &cobra.Command{
		Use:               "ca",
		Short:             "Patch CA certificates into the PIT's cloud-init meta-data.",
		DisableAutoGenTag: true,
		Long: `
Patches the Pre-Install Toolkit's (PIT) cloud-init meta-data, adding Certificate Authority (CA) certificates from a
given Shasta configuration (shasta-cfg).
`,
		Run: func(c *cobra.Command, args []string) {
			ciphertext, err := loadEncryptedCABundle(
				customizationsFile,
				sealedSecretName,
				"ca_bundle.crt",
			)
			if err != nil {
				log.Fatalf(
					"Unable to load CA data from sealed secret, %v \n",
					err,
				)
			}

			privKey, err := loadPrivateKey(sealedSecretsKeyFile)
			if err != nil {
				log.Fatalf(
					"Unable to load sealed secret private key, %v \n",
					err,
				)
			}

			plaintext, err := decryptCABundle(
				privKey,
				ciphertext,
			)
			if err != nil {
				log.Fatalf(
					"Unable to decrypt CA bundle, %v \n",
					err,
				)
			}

			CABundle, err := formatCABundle(plaintext)
			if err != nil {
				log.Fatalf(
					"Unable to format CA bundle for cloud-init, %v \n",
					err,
				)
			}

			if len(CABundle.Trusted) <= 0 {
				log.Fatalf("No CA certificates were found.")
			}

			var cloudInit Global
			var metaData Metadata
			metaData.CACerts = CABundle
			cloudInit.Global.Metadata = metaData
			update, err := json.Marshal(cloudInit)
			if err != nil {
				log.Fatalf(
					"Unable to marshal ca-certs data into JSON, %v \n",
					err,
				)
			}

			data, err := backupCloudInitData()
			if err != nil {
				log.Fatalf(
					"Failed to write backup file, %v \n",
					err,
				)
			}

			if err := writeCloudInit(
				data,
				update,
			); err != nil {
				log.Fatalf(
					"Unable to patch cloud-init seed data in place, %v \n",
					err,
				)
			}
			log.Println("Patched cloud-init seed data in place")
		},
	}
	c.Flags().StringVarP(
		&customizationsFile,
		"customizations-file",
		"",
		"",
		"path to customizations.yaml (shasta-cfg)",
	)
	c.Flags().StringVarP(
		&sealedSecretsKeyFile,
		"sealed-secret-key-file",
		"",
		"",
		"Path to sealed secrets/shasta-cfg private key",
	)
	c.Flags().StringVarP(
		&sealedSecretName,
		"sealed-secret-name",
		"",
		"gen_platform_ca_1",
		"Path to cloud-init metadata seed file",
	)
	err := c.MarkFlagRequired("customizations-file")

	if err != nil {
		log.Fatalf(
			"Failed to mark flag as required because %v",
			err,
		)
		return nil
	}
	err = c.MarkFlagRequired("sealed-secret-key-file")

	if err != nil {
		log.Fatalf(
			"Failed to mark flag as required because %v",
			err,
		)
		return nil
	}
	return c
}

// Load shasta-cfg customizations, then attempt to return the encrypted CA bundle
// from secretName -> BundleName
func loadEncryptedCABundle(
	filePath string, secretName string, bundleName string,
) (
	[]byte, error,
) {

	customizations, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var data Customizations
	if err := yaml.Unmarshal(
		customizations,
		&data,
	); err != nil {
		return nil, err
	}

	// verify sealed secret scope is cluster-wide
	clusterWide, ok := data.Spec.Kubernetes.SealedSecrets[secretName].Spec.Template.Metadata.Annotations["sealedsecrets.bitnami.com/cluster-wide"]

	// CMS is currently only using cluster-wide sealed secrets,
	// this is important as namespaced secrets include the
	// namespace as part of encryption process.
	if !ok || clusterWide != "true" {
		return nil, errors.New("sealed secret does not have cluster-wide scope, namespaced scope decryption not implemented")
	}

	b64Bundle, ok := data.Spec.Kubernetes.SealedSecrets[secretName].Spec.EncryptedData[bundleName]
	if !ok {
		return nil, errors.New("sealed secret or data attribute does not exist")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(b64Bundle)
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}

// Load and return the sealed secret RSA private key
func loadPrivateKey(filePath string) (
	*rsa.PrivateKey, error,
) {

	key, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	KeyPEM, _ := pem.Decode(key)

	privKeyParse, err := x509.ParsePKCS8PrivateKey(KeyPEM.Bytes)
	if err != nil {
		return nil, err
	}

	privKey := privKeyParse.(*rsa.PrivateKey)
	return privKey, nil
}

// Given an encrypted CA bundle from sealed secrets (ciphertext),
// decrypt and return the plaintext
func decryptCABundle(
	privKey *rsa.PrivateKey, ciphertext []byte,
) (
	[]byte, error,
) {

	// Based on https://github.com/bitnami-labs/sealed-secrets/blob/master/pkg/crypto/crypto.go

	// The first two bytes contain the length of the encrypted
	// AES session key
	if len(ciphertext) < 2 {
		return nil, errors.New("truncated ciphertext, corrupt data")
	}

	// Get the RSA encrypted AES session key length,
	// and then right shift the ciphertext
	sessionKeyLen := int(binary.BigEndian.Uint16(ciphertext))
	ciphertext = ciphertext[2:]

	if len(ciphertext) < sessionKeyLen {
		return nil, errors.New("ciphertext not long enough to hold session key, corrupt data")
	}

	// Get the RSA encrypted AES session key,
	// then right shift the ciphertext
	sessionKeyEncrypted := ciphertext[:sessionKeyLen]
	ciphertext = ciphertext[sessionKeyLen:]

	var label []byte // namespace-based scope not implemented, label is empty
	rnd := rand.Reader

	// Decrypt the AES session key
	aesSessionKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rnd,
		privKey,
		sessionKeyEncrypted,
		label,
	)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(aesSessionKey)
	if err != nil {
		return nil, err
	}

	aed, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Sealed Secrets use a zero Nonce, do the same
	zeroNonce := make(
		[]byte,
		aed.NonceSize(),
	)

	plaintext, err := aed.Open(
		nil,
		zeroNonce,
		ciphertext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Given a decrypted PEM bundle, populate and return appropriate
// cloud-init structure
func formatCABundle(raw []byte) (
	CABundle CACerts, err error,
) {

	CABundle.RemoveDefaults = false

	// parse each cert from bundle
	// convert it to a string, replace
	// newlines with \n literals, and add
	// to CloudInit data
	for {
		block, other := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {

			certPem := new(bytes.Buffer)
			err = pem.Encode(
				certPem,
				&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: block.Bytes,
				},
			)
			if err != nil {
				return CABundle, err
			}
			CABundle.Trusted = append(
				CABundle.Trusted,
				certPem.String(),
			)
		}
		raw = other
	}

	return CABundle, err
}
