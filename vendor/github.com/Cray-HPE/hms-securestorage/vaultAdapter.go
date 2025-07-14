// MIT License
//
// (C) Copyright [2019-2022] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package securestorage

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/mitchellh/mapstructure"
)

const DefaultBasePath = "secret"

// These Env var are provided globally to pods
const EnvVaultJWTFile = "CRAY_VAULT_JWT_FILE"
const EnvVaultRoleFile = "CRAY_VAULT_ROLE_FILE"
const EnvVaultAuthPath = "CRAY_VAULT_AUTH_PATH"

// Other Env vars provided by vault (https://github.com/hashicorp/vault/blob/master/api/client.go)
//
// const EnvVaultAddress = "VAULT_ADDR"
// const EnvVaultAgentAddr = "VAULT_AGENT_ADDR"
// const EnvVaultCACert = "VAULT_CACERT"
// const EnvVaultCAPath = "VAULT_CAPATH"
// const EnvVaultClientCert = "VAULT_CLIENT_CERT"
// const EnvVaultClientKey = "VAULT_CLIENT_KEY"
// const EnvVaultClientTimeout = "VAULT_CLIENT_TIMEOUT"
// const EnvVaultSkipVerify = "VAULT_SKIP_VERIFY"
// const EnvVaultNamespace = "VAULT_NAMESPACE"
// const EnvVaultTLSServerName = "VAULT_TLS_SERVER_NAME"
// const EnvVaultWrapTTL = "VAULT_WRAP_TTL"
// const EnvVaultMaxRetries = "VAULT_MAX_RETRIES"
// const EnvVaultToken = "VAULT_TOKEN"
// const EnvVaultMFA = "VAULT_MFA"
// const EnvRateLimit = "VAULT_RATE_LIMIT"
//
// Add any or all of these to your helm chart to specifiy non-default values.
// For communicating with cray's vault instance, you'll need at minimum:
// VAULT_ADDR="http://cray-vault.vault:8200"
// VAULT_SKIP_VERIFY="true"

type VaultAdapter struct {
	Config     *api.Config
	Client     VaultApi
	AuthConfig *AuthConfig
	BasePath   string
	VaultRetry int
	Role       string
}

func NewVaultAdapterAs(basePath string, role string) (SecureStorage, error) {
	ss := &VaultAdapter{
		BasePath:   basePath,
		VaultRetry: 1,
		Role: role,
	}

	// Get k8s authentication configuration values.
	authConfig := DefaultAuthConfig()
	err := authConfig.ReadEnvironment()
	if err != nil {
		return ss, err
	}

	ss.AuthConfig = authConfig

	// Get configuration values provided by the vault api
	config := api.DefaultConfig()
	err = config.ReadEnvironment()
	if err != nil {
		return ss, err
	}

	ss.Config = config

	// Create our http client for our vault connection
	client, err := api.NewClient(config)
	if err != nil {
		return ss, err
	}

	ss.Client = NewRealVaultApi(client)

	// Connect to and authenticate with vault
	err = ss.loadToken()
	if err != nil {
		return ss, err
	}

	return ss, nil
}

// Create a new SecureStorage interface that uses Vault. This connects to
// vault.
func NewVaultAdapter(basePath string) (SecureStorage, error) {
	return NewVaultAdapterAs(basePath, "")
}

// Parse an error into the vault api's ErrorResponse struct.
func getError(err error) *api.ErrorResponse {
	parsedErr := &api.ErrorResponse{}
	err = mapstructure.Decode(err, parsedErr)
	return parsedErr
}

// LoadToken loads jwt/role files from disk and attempts to generate a vault
// access token.
func (ss *VaultAdapter) loadToken() error {
	// Reload values from disk
	err := ss.AuthConfig.LoadRole()
	if err != nil {
		return err
	}

	err = ss.AuthConfig.LoadJWT()
	if err != nil {
		return err
	}

	// We will write this payload to a special auth endpoint
	k8AuthPath := ss.AuthConfig.GetAuthPath()
	k8AuthArgs := ss.AuthConfig.GetAuthArgs()

	// Apply role override if any
	if ss.Role != "" {
		k8AuthArgs["role"] = ss.Role
	}

	secret, err := ss.Client.Write(k8AuthPath, k8AuthArgs)
	if err != nil {
		return err
	}
	tokenID, err := secret.TokenID()
	if err != nil {
		return err
	}

	ss.Client.SetToken(tokenID)
	return nil
}

func (ss *VaultAdapter) checkErrForTokenRefresh(err error) bool {
	lowerErrorString := strings.ToLower(err.Error())

	if strings.Contains(lowerErrorString, "code: 403") ||
		strings.Contains(lowerErrorString, "missing client token") {
		return true
	}

	return false
}

// Write a struct to Vault at the location specified by key. This function
// prepends the basePath. Retries are implemented for token renewal.
func (ss *VaultAdapter) Store(key string, value interface{}) error {
	var (
		err  error
		data map[string]interface{}
	)

	err = mapstructure.Decode(value, &data)
	if err != nil {
		return err
	}
	path := ss.BasePath + "/" + key
	for i := 0; i <= ss.VaultRetry; i++ {
		// Write the data to Vault
		_, err = ss.Client.Write(path, data)
		if err != nil {
			if ss.checkErrForTokenRefresh(err) {
				// We need to renew the token and then retry
				if err = ss.loadToken(); err != nil {
					return err
				} else {
					continue
				}
			} else {
				return err
			}
		}
		break
	}
	return err
}

// Write a struct to Vault at the location specified by key and return the response.
// This function prepends the basePath. Retries are implemented for token renewal.
// Note: Unlike Lookup(), this returns the entire response body. Not just secretValues.Data.
func (ss *VaultAdapter) StoreWithData(key string, value interface{}, output interface{}) error {
	var (
		err  error
		data map[string]interface{}
	)

	err = mapstructure.Decode(value, &data)
	if err != nil {
		return err
	}
	path := ss.BasePath + "/" + key
	for i := 0; i <= ss.VaultRetry; i++ {
		// Write the data to Vault
		secretValues, err := ss.Client.Write(path, data)
		if err != nil {
			if ss.checkErrForTokenRefresh(err) {
				// We need to renew the token and then retry
				if err = ss.loadToken(); err != nil {
					return err
				} else {
					continue
				}
			} else {

				return err
			}
		}

		if secretValues == nil {
			// No data returned. 
			break
		}

		err = mapstructure.Decode(secretValues, output)
		break
	}

	return err
}

// Read a struct from Vault at the location specified by key. This function
// prepends the basePath. Retries are implemented for token renewal.
func (ss *VaultAdapter) Lookup(key string, output interface{}) error {
	var err error

	if output == nil {
		return fmt.Errorf("output interface was nil")
	}
	path := ss.BasePath + "/" + key
	for i := 0; i <= ss.VaultRetry; i++ {
		// Read the data from Vault
		secretValues, err := ss.Client.Read(path)
		if err != nil {
			if ss.checkErrForTokenRefresh(err) {
				// We need to renew the token and then retry
				if err = ss.loadToken(); err != nil {
					return err
				} else {
					continue
				}
			} else {
				return err
			}
		}

		if secretValues == nil {
			// Not considering this an error as the Read technically
			// worked, there just wasn't anything there.
			break
		}

		err = mapstructure.Decode(secretValues.Data, output)
		break
	}

	return err
}

// Remove a struct from Vault at the location specified by key. This function
// prepends the basePath. Retries are implemented for token renewal.
func (ss *VaultAdapter) Delete(key string) error {
	var err error

	path := ss.BasePath + "/" + key
	for i := 0; i <= ss.VaultRetry; i++ {
		// Remove the key and data from Vault
		_, err := ss.Client.Delete(path)
		if err != nil {
			if ss.checkErrForTokenRefresh(err) {
				// We need to renew the token and then retry
				if err = ss.loadToken(); err != nil {
					return err
				} else {
					continue
				}
			} else {
				return err
			}
		}
		break
	}

	return err
}

// Get a list of keys that exsist in Vault at the path specified by keyPath.
// This function prepends the basePath. Retries are implemented for token
// renewal.
func (ss *VaultAdapter) LookupKeys(keyPath string) ([]string, error) {
	var (
		err   error
		klist []string
	)

	path := ss.BasePath + "/" + keyPath
	for i := 0; i <= ss.VaultRetry; i++ {
		secretValues, err := ss.Client.List(path)
		if err != nil {
			if ss.checkErrForTokenRefresh(err) {
				// We need to renew the token and then retry
				if err = ss.loadToken(); err != nil {
					return nil, err
				} else {
					continue
				}
			} else {
				return nil, err
			}
		}
		keys, ok := secretValues.Data["keys"].([]interface{})
		if !ok {
			return klist, fmt.Errorf("Cannot get secret data")
		}
		for _, key := range keys {
			xname, ok := key.(string)
			if !ok {
				return klist, fmt.Errorf("Cannot make key into string")
			}
			klist = append(klist, xname)
		}
		break
	}

	return klist, err
}

///////////////////////////////
// K8s Authentication functions
///////////////////////////////

// AuthConfig struct for vault k8s authentication
type AuthConfig struct {
	JWTFile  string
	RoleFile string
	Path     string
	jwt      string
	role     string
}

// DefaultAuthConfig Create the default auth config that will work for almost all scenarios
func DefaultAuthConfig() *AuthConfig {
	authConfig := &AuthConfig{
		JWTFile:  "/var/run/secrets/kubernetes.io/serviceaccount/token",
		RoleFile: "/var/run/secrets/kubernetes.io/serviceaccount/namespace",
		Path:     "auth/kubernetes/login",
	}

	return authConfig
}

// ReadEnvironment Update an authConfig with environment variables
func (authConfig *AuthConfig) ReadEnvironment() error {
	var jwtFile string
	var roleFile string
	var authPath string

	if v := os.Getenv(EnvVaultJWTFile); v != "" {
		jwtFile = v
	}
	if v := os.Getenv(EnvVaultRoleFile); v != "" {
		roleFile = v
	}
	if v := os.Getenv(EnvVaultAuthPath); v != "" {
		authPath = v
	}

	if jwtFile != "" {
		authConfig.JWTFile = jwtFile
	}
	if roleFile != "" {
		authConfig.RoleFile = roleFile
	}
	if authPath != "" {
		authConfig.Path = authPath
	}

	return nil
}

func getFileContents(filePath string) (string, error) {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(contents)), nil
}

// LoadJWT save contents of JWTFile to the jwt field
func (authConfig *AuthConfig) LoadJWT() error {
	jwt, err := getFileContents(authConfig.JWTFile)
	if err != nil {
		return err
	}
	authConfig.jwt = jwt
	return nil
}

// LoadRole save contents of RoleFile to the role field
func (authConfig *AuthConfig) LoadRole() error {
	role, err := getFileContents(authConfig.RoleFile)
	if err != nil {
		return err
	}
	authConfig.role = role
	return nil
}

// GetAuthPath Getter for auth path key
func (authConfig *AuthConfig) GetAuthPath() string {
	return authConfig.Path
}

// GetAuthArgs generates the ars required for generating an auth token
func (authConfig *AuthConfig) GetAuthArgs() map[string]interface{} {
	authArgs := map[string]interface{}{
		"role": authConfig.role,
		"jwt":  authConfig.jwt,
	}
	return authArgs
}

///////////////////////////////////////////////////////////////////////////////
// Vault API interface - This interface wraps only a subset of functions for
// api.Client so as to reduce the amount of functions that need to be mocked
// for unit testing.
///////////////////////////////////////////////////////////////////////////////
type VaultApi interface {
	Read(path string) (*api.Secret, error)
	Write(path string, data map[string]interface{}) (*api.Secret, error)
	Delete(path string) (*api.Secret, error)
	List(path string) (*api.Secret, error)
	SetToken(t string)
}

type RealVaultApi struct {
	Client *api.Client
}

func NewRealVaultApi(client *api.Client) VaultApi {
	v := &RealVaultApi{
		Client: client,
	}
	return v
}

func (v *RealVaultApi) Read(path string) (*api.Secret, error) {
	return v.Client.Logical().Read(path)
}

func (v *RealVaultApi) Write(path string, data map[string]interface{}) (*api.Secret, error) {
	return v.Client.Logical().Write(path, data)
}

func (v *RealVaultApi) Delete(path string) (*api.Secret, error) {
	return v.Client.Logical().Delete(path)
}

func (v *RealVaultApi) List(path string) (*api.Secret, error) {
	return v.Client.Logical().List(path)
}

func (v *RealVaultApi) SetToken(t string) {
	v.Client.SetToken(t)
}
