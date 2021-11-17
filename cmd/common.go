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
