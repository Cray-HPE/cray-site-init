/*
 MIT License

 (C) Copyright 2025 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package csm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	csiKubernetes "github.com/Cray-HPE/cray-site-init/pkg/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DefaultBaseAPIURL is the URL for which all API requests should be directed at.
const DefaultBaseAPIURL = "https://api-gw-service-nmn.local"

// DefaultAdminTokenSecretNamespace is the default namespace to use when interfacing with Kubernetes.
const DefaultAdminTokenSecretNamespace = "default"

// DefaultAdminTokenSecretName is the default name of the secret containing the OpenID authentication information for CSM APIs.
const DefaultAdminTokenSecretName = "admin-client-auth"

type credentials struct {
	AccessToken string `json:"access_token"`
}

var (
	AdminTokenSecretName      = DefaultAdminTokenSecretName
	AdminTokenSecretNamespace = DefaultAdminTokenSecretNamespace
	BaseAPIURL                = DefaultBaseAPIURL
)

// GetToken returns an API token for communicating with CSM's various APIs.
func GetToken() (token string, err error) {

	kc, err := csiKubernetes.NewKubernetesClientRaw()
	if err != nil {
		return token, fmt.Errorf(
			"error creating Kubernetes client: %v",
			err,
		)
	}
	secret, err := getSecret(
		kc,
		AdminTokenSecretNamespace,
		AdminTokenSecretName,
	)
	if err != nil {
		return token, fmt.Errorf(
			"error getting OpenID secret (%s/%s) because %v",
			AdminTokenSecretNamespace,
			AdminTokenSecretName,
			err,
		)
	}
	bearer, err := requestBearer(secret)
	var creds credentials
	err = json.Unmarshal(
		bearer,
		&creds,
	)

	token = creds.AccessToken
	return token, err
}

func getSecret(client *kubernetes.Clientset, namespace string, secretName string) (token map[string][]byte, err error) {
	secret, err := client.CoreV1().Secrets(namespace).Get(
		context.Background(),
		secretName,
		metav1.GetOptions{},
	)
	if err != nil {
		return nil, err
	}
	return secret.Data, nil
}

func requestBearer(secret map[string][]byte) (responseBody []byte, err error) {
	clientIDName := "client-id"
	clientSecretName := "client-secret"
	endpointName := "endpoint"
	clientID := string(secret[clientIDName])
	encodedClientSecret := string(secret[clientSecretName])
	grantType := "client_credentials"

	var endpoint string
	endpoint = string(secret[endpointName])

	if !strings.HasPrefix(
		endpoint,
		BaseAPIURL,
	) || BaseAPIURL != DefaultBaseAPIURL {
		endpointURL, _ := url.Parse(endpoint)
		endpoint = fmt.Sprintf(
			"%s%s",
			BaseAPIURL,
			endpointURL.Path,
		)
	}

	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {encodedClientSecret},
		"grant_type":    {grantType},
	}

	req, err := http.NewRequest(
		"POST",
		endpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		log.Fatalf(
			"Failed to create an HTTP client for accessing the API gateway:\n%v",
			err,
		)
	}
	req.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf(
			"Failed to do request:\n%v",
			err,
		)
	}

	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf(
			"Failed to decode API gateway response:\n%v",
			err,
		)
	}
	defer resp.Body.Close()

	return responseBody, err
}
