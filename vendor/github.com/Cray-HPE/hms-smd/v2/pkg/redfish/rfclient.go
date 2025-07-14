// MIT License
//
// (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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

package rf

import (
	"os"

	"github.com/Cray-HPE/hms-certs/pkg/hms_certs"
)

var httpRFClient *hms_certs.HTTPClientPair
var httpClientTimeout = 30

//var httpClientProxyURL = ""
//var httpClientInsecureSkipVerify = true

// Setter functions for above.

// Set the HTTP Client timeout in seconds used during Redfish interogation.
// 0 means no timeout.
// NOTE: Global, to be called only once at startup.
func SetHTTPClientTimeout(timeout int) {
	if timeout >= 0 {
		httpClientTimeout = timeout
	} else {
		errlog.Printf("SetHTTPClientTimeout: bad arg '%d'", timeout)
	}
}

// Get the HTTP Client timeout in seconds used during Redfish interogation
func GetHTTPClientTimeout() int {
	return httpClientTimeout
}

/*
// Set HTTP client proxy used during Redfish interogation, including port
// and protocol (see http package: socks5, http, https).  Defaults assigned
// if info is missing.  If unparsable, will default to no proxy.
// NOTE: Global, to be called only once at startup.
func SetHTTPClientProxyURL(proxyURLStr string) {
	httpClientProxyURL = proxyURLStr
}

// Get HTTP client proxy used during Redfish interogation
func GetHTTPClientProxyURL() string {
	return httpClientProxyURL
}

// Set HTTP client InsecureSkipVerify flag used during Redfish interogation.
func SetHTTPClientInsecureSkipVerify(flag bool) {
	httpClientInsecureSkipVerify = flag
}

// Get HTTP client InsecureSkipVerify flag used during Redfish interogation.
func GetHTTPClientInsecureSkipVerify() bool {
	return httpClientInsecureSkipVerify
}
*/

// Returns default-configuration HTTP Client
func RfDefaultClient() *hms_certs.HTTPClientPair {
	var cerr error
	if httpRFClient == nil {
		uri := os.Getenv("SMD_CA_URI")
		// TODO: Why CreateHTTPClientPair() instead of CreateRetryableHTTPClientPair()??
		httpRFClient, cerr = hms_certs.CreateHTTPClientPair(uri, httpClientTimeout)
		if cerr != nil {
			errlog.Printf("Can't create TLS cert-enabled HTTP transport, reverting to less secure transport.")
			httpRFClient, cerr = hms_certs.CreateHTTPClientPair("", httpClientTimeout)
			if cerr != nil {
				errlog.Printf("Can't create any HTTP transport!")
				httpRFClient = nil
				return nil
			}
		}
	}
	return httpRFClient
}

/*
// Returns default-configuration HTTP Client with proxy.  If invalid
// proxy string given, no proxy will be used.
// TODO: Need to have a way to specify the CA cert used to verify
//       the Redfish endpoint
func RfProxyClient(proxyURLStr string) http.Client {
	proxyURL, err := url.Parse(proxyURLStr)
	if err != nil {
		errlog.Printf("Can't parse '%s', not using proxy: %s",
			proxyURLStr, err)
		return RfDefaultClient()
	}
	timeout := time.Duration(httpClientTimeout) * time.Second
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: httpClientInsecureSkipVerify,
		},
		Proxy: http.ProxyURL(proxyURL),
	}
	client := http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client
}
*/
