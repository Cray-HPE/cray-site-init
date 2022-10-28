/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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

package cmd

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
)

// Common vars.
var (
	httpClient *http.Client

	kubeconfig string

	bssBaseURL string
	hsmBaseURL string
	slsBaseURL string

	token string
)

func checkToken() {
	if token == "" {
		log.Panicln("Environment variable TOKEN can NOT be blank!")
	}
}

func setupEnvs() {
	token = os.Getenv("TOKEN")

	bssBaseURL = os.Getenv("BSS_BASE_URL")
	if bssBaseURL == "" {
		bssBaseURL = "https://api-gw-service-nmn.local/apis/bss"
		checkToken()
	}

	hsmBaseURL = os.Getenv("HSM_BASE_URL")
	if hsmBaseURL == "" {
		hsmBaseURL = "https://api-gw-service-nmn.local/apis/smd"
		checkToken()
	}

	slsBaseURL = os.Getenv("SLS_BASE_URL")
	if slsBaseURL == "" {
		slsBaseURL = "https://api-gw-service-nmn.local/apis/sls"
		checkToken()
	}
}

func setupHTTPClient() {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	httpClient = &http.Client{Transport: transport}
}
