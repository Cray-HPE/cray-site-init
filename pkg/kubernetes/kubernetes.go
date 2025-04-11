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

package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

// DefaultKubeConfig is the default location for the Kubernetes config.
const DefaultKubeConfig = "~/.kube/config"

// DefaultKubeConfigEnvVar is the default environment variable to read that contains the path to the user's active Kubernetes config.
const DefaultKubeConfigEnvVar = "KUBECONFIG"

func findFallbackConfig() (path string) {
	fallbackConfig := os.Getenv(DefaultKubeConfigEnvVar)
	if fallbackConfig == "" {
		fallbackConfig = DefaultKubeConfig
	}
	return fallbackConfig
}

// NewKubernetesClientRaw returns a no-frills, Kubernetes client.
func NewKubernetesClientRaw() (client *kubernetes.Clientset, err error) {
	kubeConfig := findFallbackConfig()
	config, configErr := clientcmd.BuildConfigFromFlags(
		"",
		kubeConfig,
	)
	if configErr != nil {
		err = fmt.Errorf(
			"failed to build config from kubeconfig file: %w",
			configErr,
		)
		return
	}

	clientSet, clientsetErr := kubernetes.NewForConfig(config)
	if clientsetErr != nil {
		err = fmt.Errorf(
			"failed to create clientset: %w",
			clientsetErr,
		)
		return
	}
	return clientSet, nil
}
