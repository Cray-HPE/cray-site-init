package cmd

/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

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
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/spf13/cobra"

	jsonpatch "github.com/evanphx/json-patch"
	"gopkg.in/yaml.v2"
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
	apiVersion string
	Spec       struct {
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

// CloudInitGlobal - Boilerplace for cloud-init hierarchical structure
type CloudInitGlobal struct {
	Global struct {
		Metadata `json:"meta-data"`
	}
}

var customizationsFile string
var sealedSecretsKeyFile string
var cloudInitSeedFile string
var sealedSecretName string

var patchCA = &cobra.Command{
	Use:   "ca",
	Short: "Patch cloud-init metadata with CA certs",
	Long: `
Patch cloud-init metadata (in place) with certificate authority (CA) certificates from
shasta-cfg (customizations.yaml). Decrypts CA material from named sealed secret using the shasta-cfg
private RSA key.`,
	Run: func(cmd *cobra.Command, args []string) {
		ciphertext, err := loadEncryptedCABundle(customizationsFile, sealedSecretName, "ca_bundle.crt")
		if err != nil {
			log.Fatalf("Unable to load CA data from sealed secret, %v \n", err)
		}

		privKey, err := loadPrivateKey(sealedSecretsKeyFile)
		if err != nil {
			log.Fatalf("Unable to load sealed secret private key, %v \n", err)
		}

		plaintext, err := decryptCABundle(privKey, ciphertext)
		if err != nil {
			log.Fatalf("Unable to decrypt CA bundle, %v \n", err)
		}

		CABundle, err := formatCABundle(plaintext)
		if err != nil {
			log.Fatalf("Unable to format CA bundle for cloud-init, %v \n", err)
		}

		if len(CABundle.Trusted) <= 0 {
			log.Fatalf("No CA certificates were found.")
		}

		var cloudInit CloudInitGlobal
		cloudInit.Global.Metadata.CACerts = CABundle
		update, err := json.Marshal(cloudInit)
		if err != nil {
			log.Fatalf("Unable to marshal ca-certs data into JSON, %v \n", err)
		}

		data, err := ioutil.ReadFile(cloudInitSeedFile)
		if err != nil {
			log.Fatalf("Unable to load cloud-init seed data, %v \n", err)
		}
		merged, err := jsonpatch.MergePatch(data, update)
		if err != nil {
			log.Fatalf("Could not create merge patch to update cloud-init seed data, %v \n", err)
		}

		// write original cloud-init data to backup
		currentTime := time.Now()
		ts := currentTime.Unix()
		backupFile := cloudInitSeedFile + "-" + fmt.Sprintf("%d", ts)
		err = ioutil.WriteFile(backupFile, data, 0640)
		if err != nil {
			log.Fatalf("Unable to create backup of cloud-init seed data, %v \n", err)
		}
		log.Println("Backup of cloud-init seed data at", backupFile)

		// Unmarshal merged cloud-init data, marshal it back with indent
		// then write it to the original cloud-init file (in place patch)
		var mergeUnmarshal map[string]interface{}
		json.Unmarshal(merged, &mergeUnmarshal)
		mergeMarshal, _ := json.MarshalIndent(mergeUnmarshal, "", "  ")
		err = ioutil.WriteFile(cloudInitSeedFile, mergeMarshal, 0640)
		if err != nil {
			log.Fatalf("Unable to patch cloud-init seed data in place, %v \n", err)
		}
		log.Println("Patched cloud-init seed data in place")
	},
}

func init() {
	patchCmd.AddCommand(patchCA)
	patchCA.DisableAutoGenTag = true
	patchCA.Flags().StringVarP(&customizationsFile, "customizations-file", "", "", "path to customizations.yaml (shasta-cfg)")
	patchCA.Flags().StringVarP(&cloudInitSeedFile, "cloud-init-seed-file", "", "", "Path to cloud-init metadata seed file")
	patchCA.Flags().StringVarP(&sealedSecretsKeyFile, "sealed-secret-key-file", "", "", "Path to sealed secrets/shasta-cfg private key")
	patchCA.Flags().StringVarP(&sealedSecretName, "sealed-secret-name", "", "gen_platform_ca_1", "Path to cloud-init metadata seed file")
	patchCA.MarkFlagRequired("customizations-file")
	patchCA.MarkFlagRequired("cloud-init-seed-file")
	patchCA.MarkFlagRequired("sealed-secret-key-file")
}

// Load shasta-cfg customizations, then attempt to return the encrypted CA bundle
// from secretName -> BundleName
func loadEncryptedCABundle(filePath string, secretName string, bundleName string) ([]byte, error) {

	customizations, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var data Customizations
	if err := yaml.Unmarshal(customizations, &data); err != nil {
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
func loadPrivateKey(filePath string) (*rsa.PrivateKey, error) {

	key, err := ioutil.ReadFile(filePath)
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
func decryptCABundle(privKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {

	// Based on https://github.com/bitnami-labs/sealed-secrets/blob/master/pkg/crypto/crypto.go

	// The first two bytes contain the length of the encrypted
	// AES session key
	if len(ciphertext) < 2 {
		return nil, errors.New("truncuated ciphertext, corrupt data?")
	}

	// Get the RSA encrypted AES session key length,
	// and then right shift the ciphertext
	sessionKeyLen := int(binary.BigEndian.Uint16(ciphertext))
	ciphertext = ciphertext[2:]

	if len(ciphertext) < sessionKeyLen {
		return nil, errors.New("ciphertext not long enough to hold session key, corrupt data?")
	}

	// Get the RSA encrypted AES session key,
	// then right shift the ciphertext
	sessionKeyEncrypted := ciphertext[:sessionKeyLen]
	ciphertext = ciphertext[sessionKeyLen:]

	var label []byte // namespace-based scope not implemented, label is empty
	rnd := rand.Reader

	// Decrypt the AES session key
	aesSessionKey, err := rsa.DecryptOAEP(sha256.New(), rnd, privKey, sessionKeyEncrypted, label)
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
	zeroNonce := make([]byte, aed.NonceSize())

	plaintext, err := aed.Open(nil, zeroNonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Given a decrypted PEM bundle, populate and return appropriate
// cloud-init structure
func formatCABundle(raw []byte) (CACerts, error) {

	var CABundle CACerts
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
			pem.Encode(certPem, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: block.Bytes,
			})
			CABundle.Trusted = append(CABundle.Trusted, certPem.String())
		}
		raw = other
	}

	return CABundle, nil
}
