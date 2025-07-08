// MIT License
//
// (C) Copyright [2020-2022,2025] Hewlett Packard Enterprise Development LP
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

package hms_certs

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	base "github.com/Cray-HPE/hms-base/v2"
	sstorage "github.com/Cray-HPE/hms-securestorage"
	"github.com/Cray-HPE/hms-xname/xnametypes"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

// This package provides a means to create TLS cert/key pairs as well as
// to fetch the CA bundle/chain.    See README.md for more info.

// Used for authentication in k8s, to get to the vault data

type vaultTokStuff struct {
	Auth vtsAuth `json:"auth"`
}

type vtsAuth struct {
	ClientToken string `json:"client_token"`
}

// Used to create certs
// hms-securestorage uses the mapstructure pkg for decoding structs into the map[string]interface{}
// type needed for the Vault API. The 'mapstructure' tag ensures that the field names are correct.
type vaultCertReq struct {
	CommonName string `json:"common_name" mapstructure:"common_name"`
	TTL        string `json:"ttl" mapstructure:"ttl"`
	AltNames   string `json:"alt_names" mapstructure:"alt_names"`
}

type VaultCertData struct {
	RequestID     string   `json:"request_id"`
	LeaseID       string   `json:"lease_id"`
	Renewable     bool     `json:"renewable"`
	LeaseDuration int      `json:"lease_duration"`
	Data          CertInfo `json:"data"`
}

type CertInfo struct {
	CAChain        []string `json:"ca_chain"`
	Certificate    string   `json:"certificate"`
	Expiration     int      `json:"expiration"`
	IssuingCA      string   `json:"issuing_ca"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType string   `json:"private_key_type"`
	SerialNumber   string   `json:"serial_number"`
	FQDN           string   `json:"fqdn,omitempty"`
}

/* This is what we get back from the Vault PKI when requesting keys
{
  "request_id": "dead1562-4a3c-6828-9951-d85d1997e0ce",
  "lease_id": "",
  "renewable": false,
  "lease_duration": 0,
  "data": {
    "ca_chain": [
      "-----BEGIN CERTIFICATE-----\nxxx\n-----END CERTIFICATE-----",
      "-----BEGIN CERTIFICATE-----\nyyy\n-----END CERTIFICATE-----"
    ],
    "certificate": "-----BEGIN CERTIFICATE-----\naaa\n-----END CERTIFICATE-----",
    "expiration": 1627423464,
    "issuing_ca": "-----BEGIN CERTIFICATE-----\nbbb\n-----END CERTIFICATE-----",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\nccc\n-----END RSA PRIVATE KEY-----",
    "private_key_type": "rsa",
    "serial_number": "4f:fe:98:c2:0d:d4:1e:bb:50:75:8b:94:fe:b9:48:89:b6:d4:7f:86"
  },
  "wrap_info": null,
  "warnings": null,
  "auth": null
}*/


// Used for storing cert info in Vault secure storage

type CertStorage struct {
	Cert string `json:"Cert"`
	Key  string `json:"Key"`
}

// HTTP client pack

type HTTPClientPair struct {
	SecureClient   *retryablehttp.Client
	InsecureClient *retryablehttp.Client
	MaxRetryCount  int
	MaxRetryWait   int
	FailedOver     bool			//true if most recent op failed over
}


// Configuration parameters. Changeable by applications, but DO NOT
// change them unless you know what you're doing!!

type Config struct {
	VaultKeyBase        string	//Defaults to vaultKeyBase
	CertKeyBasePath     string	//Defaults to certKeyBasePath
	VaultPKIBase        string	//Defaults to vaultPKIBase
	PKIPath             string	//Defaults to pkiPath
	CAChainPath         string	//Defaults to caPath
	LogInsecureFailover bool	//Defaults to true
}

// Constants available to the user of this package

const (
	CertDomainCabinet = "CERT_DOMAIN_CABINET"
	CertDomainChassis = "CERT_DOMAIN_CHASSIS"
	CertDomainBlade   = "CERT_DOMAIN_BLADE"
	CertDomainBMC     = "CERT_DOMAIN_BMC"

	VaultCAChainURI = "vault://pki_common/ca_chain"
)

// Constants used within this package

const (

	vaultKeyBase    = "secret"
	certKeyBasePath = "certs"

	vaultPKIBase    = "pki_common"
	pkiPath         = "issue/pki-common"
	caChainPath     = "ca_chain"

	maxCabChassis    = 8
	maxChassisSlot   = 8
	maxRVChassisSlot = 64
	maxSlotBMC       = 8
)


// Config params, changeable by user of this package.  See constants above.
// Note that in addition to these parameters there are some environment 
// variables which affect the way Vault works, and they are global to the
// application:
//
// CRAY_VAULT_JWT_FILE    # The file containing the access token.
// CRAY_VAULT_ROLE_FILE   # Namespace file.  Default is /var/run/secrets/kubernetes.io/serviceaccount/namespace
// CRAY_VAULT_AUTH_PATH   # Vault URL tail for k8s logins.  Default is
//                        # /auth/kubernetes/login
// VAULT_ADDR             # URL of Vault, default is http://cray-vault.vault:8200

var ConfigParams = Config{VaultKeyBase:        vaultKeyBase,
                          CertKeyBasePath:     certKeyBasePath,
                          VaultPKIBase:        vaultPKIBase,
                          PKIPath:             pkiPath,
                          CAChainPath:         caChainPath,
                          LogInsecureFailover: true,
}

// local global vars
var logger = logrus.New()
var __httpTransport *http.Transport
var __httpClient *http.Client
var __vaultEnabled = true
var cbmap  = make(map[string]bool)
var vstore = make(map[string]VaultCertData)
var instName string


// Initialize the certs package.  This pretty much just sets up the logging.

func Init(loggerP *logrus.Logger) {
    if loggerP != nil {
        logger = loggerP
    } else {
        logger = logrus.New()
    }

	//Check for Vault-ness.  No ENV var == vault is enabled, so specifically
	//look for 0, false, etc.

	ven := os.Getenv("VAULT_ENABLE")
	if ((ven == "0") || (strings.ToLower(ven) == "false")) {
		__vaultEnabled = false
	}
}

func InitInstance(loggerP *logrus.Logger, inst string) {
	instName = inst
	Init(loggerP)
}

func seclog_Errorf(fmt string, args ...interface{}) {
	if (ConfigParams.LogInsecureFailover) {
		logger.Errorf(fmt,args...)
	}
}

func seclog_Warnf(fmt string, args ...interface{}) {
	if (ConfigParams.LogInsecureFailover) {
		logger.Warnf(fmt,args...)
	}
}

// Register for changes to a CA chain.  This is based on a URI, which can be 
// a filename or VaultCAChainURI.
//
// If file, it will put a watch on the file.  If vault URI, it will poll.
// In either case, when the CA has changed, the specified function is called
// passing in the info needed to re-do ones' HTTP connection.
//
// uri(in):  CA chain resource name.  Can be a full pathname (used with
//           configmaps) or the vault URI for the CAChain (VaultCAChainURI).
// cb(in):   Function to call when the CA chain resource changes.  The function
//           must take a string as an argument; this string is the new CA
//           chain data, which can then be used to re-do HTTP transports.
// Return:   nil on success, error info on error.

func CAUpdateRegister(uri string, cb func(string)) error {
	if (uri == VaultCAChainURI) {
		baseChain,berr := FetchCAChain(uri)
		if (berr != nil) {
			return berr
		}

		cbmap[uri] = true
		go func() {
			for {
				time.Sleep(10 * time.Second)
				if (cbmap[uri] == false) {
					break
				}
				newChain,nerr := FetchCAChain(uri)
				if (nerr != nil) {
					logger.Errorf("%v",nerr)
					continue
				}
				if (baseChain != newChain) {
					baseChain = newChain
					cb(newChain)
				}
			}
		}()
	} else {
		finfo,ferr := os.Stat(uri)
		if (ferr != nil) {
			return fmt.Errorf("Error stat-ing file '%s': %v",
				uri,ferr)
		}
		orgTime := finfo.ModTime()
		cbmap[uri] = true

		go func() {
			for {
				time.Sleep(10 * time.Second)
				if (cbmap[uri] == false) {
					break
				}
				ninfo,nerr := os.Stat(uri)
				if (nerr != nil) {
					logger.Errorf("Error stat-ing file '%s': %v",
						uri,ferr)
					continue
				}
				if (ninfo.ModTime() != orgTime) {
					orgTime = ninfo.ModTime()
					chain,cerr := ioutil.ReadFile(uri)
					if (cerr != nil) {
						logger.Errorf("Error reading CA chain file '%s': %v",
							uri,cerr)
					} else {
						cb(string(chain))
					}
				}
			}
		}()
	}

	return nil
}

// Un-register a CA chain change callback function.
//
// uri(in):  CA chain resource name.  Can be a full pathname (used with
//           configmaps) or the vault URI for the CAChain (VaultCAChainURI).
//           This is the "key" used to identify the registration.
// Return:   nil on success, error info on error.

func CAUpdateUnregister(uri string) error {
	ok,_ := cbmap[uri]
	if (!ok) {
		return fmt.Errorf("No CA chain callback map entry found for '%s'",uri)
	}
	delete(cbmap,uri)
	return nil
}

// Given a raw key, massage it into a proper vault key (prepend path).

func vaultKey(raw string) string {
	return path.Join(ConfigParams.CertKeyBasePath,raw)
}

// Given an endpoint and a domain type, generate all possible SANs for a cert.
//
// endpoint(in): XName of an endpoint in a cert domain, e.g., "x1000" for a
//               cabinet domain.  Can be any xname in that cabinet.
// domain(in):   Domain ID, e.g., CertDomainCabinet.
// Return:       Comma separated string containing all generated SANs;
//               nil on success; error info on error.

func genAllDomainAltNames(endpoint,domain string) (string,error) {
	var eps,toks []string
	var chassis,slot,bmc,cid,maxSlot int

	switch (domain) {
		case CertDomainCabinet:
			//Create all node cards, switch cards, CMMs, and PDUs.
			//This has to handle rtrs too and c0 has to have up to 64 rtrs
			//and blades for river
			eps = append(eps,fmt.Sprintf("%sc0",endpoint))
			for slot = 0; slot < 4; slot ++ {
				eps = append(eps,fmt.Sprintf("%sm%d",endpoint,slot))
				eps = append(eps,fmt.Sprintf("%sm%d-rts",endpoint,slot))
			}
			for slot = 0; slot < maxRVChassisSlot; slot ++ {
				for bmc = 0; bmc < maxSlotBMC; bmc ++ {
					eps = append(eps,fmt.Sprintf("%sc%ds%db%d",
							endpoint,0,slot,bmc))
					eps = append(eps,fmt.Sprintf("%sc%dr%db%d",
							endpoint,0,slot,bmc))
				}
			}
			for chassis = 1; chassis < maxCabChassis; chassis ++ {
				eps = append(eps,fmt.Sprintf("%sc%d",endpoint,chassis))
				for slot = 0; slot < maxChassisSlot; slot ++ {
					for bmc = 0; bmc < maxSlotBMC; bmc ++ {
						eps = append(eps,fmt.Sprintf("%sc%ds%db%d",
								endpoint,chassis,slot,bmc))
						eps = append(eps,fmt.Sprintf("%sc%dr%db%d",
								endpoint,chassis,slot,bmc))
					}
				}
			}
			break

		case CertDomainChassis:
			//If this is c0, slots and rtrs, up to 64
			toks = strings.Split(endpoint,"c")
			if (len(toks) < 2) {
				return "",fmt.Errorf("Invalid chassis name: '%s' (missing 'c')",
							endpoint)
			}
			cid,_ = strconv.Atoi(toks[1])
			if (cid == 0) {
				maxSlot = maxRVChassisSlot
			} else {
				maxSlot = maxChassisSlot
			}
			for slot = 0; slot < maxSlot; slot ++ {
				for bmc = 0; bmc < maxSlotBMC; bmc ++ {
					eps = append(eps,fmt.Sprintf("%ss%db%d",
							endpoint,slot,bmc))
					eps = append(eps,fmt.Sprintf("%sr%db%d",
							endpoint,slot,bmc))
				}
			}
			break

		case CertDomainBlade:
			//Note: endpoint will be either xXcCsS or xXcCrR
			for bmc = 0; bmc < maxSlotBMC; bmc ++ {
				eps = append(eps,fmt.Sprintf("%sb%d",
						endpoint,bmc))
			}
			break

		case CertDomainBMC:
			toks = strings.Split(endpoint,"n")
			eps = append(eps,toks[0])
			break

		default:
			return "",fmt.Errorf("Invalid cert domain: %s",domain)
	}

	return strings.Join(eps,","),nil
}

// Given an XName and a separator, get the front part of an XName
//
// xname(in): Full xname e.g. x1000c1s2b0
// sep(in):   Separator, e.g. "c"
// Return:    Front part of xname, e.g., "x1000"

func getXNameSegment(xname,sep string) string {
	if (sep == "") {
		return xname
	}
	toks := strings.Split(xname,sep)
	if (len(toks) < 2) {
		return xname
	}
	return toks[0]
}

// Given an array of endpoints and a domain type, check all endpoints to be
// sure they are all in the same domain.  Return the domain XName.
//
// endpoints(in): Array of BMC XNames
// domain(in):    Domain type, e.g. CertDomainCabinet
// sep(in):       XName separator, e.g. "c"
// Return:        Domain XName, e.g. "x1000"
//                nil on success, error info on error.

func checkDomainTargs(endpoints []string, domain string, sep string) (string,error) {
	toks := strings.Split(endpoints[0],":")
	xname := getXNameSegment(toks[0],sep)
	for ix := 0; ix < len(endpoints); ix ++ {
		ttoks := strings.Split(endpoints[ix],":")
		compName := getXNameSegment(ttoks[0],sep)
		if (compName != xname) {
			return "",fmt.Errorf("ERROR, endpoint not in %s domain: %s",
				domain,endpoints[ix])
		}
		//There can be -xxx annotations in some cases, e.g. x0m0-rts, so strip
		//off anything with a dash.
		dtoks := strings.Split(ttoks[0],"-")
		if (xnametypes.VerifyNormalizeCompID(dtoks[0]) == "") {
			return "",fmt.Errorf("ERROR, endpoint not a valid XName: %s (%s)",
				ttoks[0],endpoints[ix])
		}
	}

	return xname,nil
}

// Given a list of BMC endpoints and a domain type, verify that all endpoints 
// are contained in the same cert domain and return the domain xname.
//
// endpoints(in): Array of BMC XNames
// domain(in):    Domain type, e.g. CertDomainCabinet
// Return:        Domain XName, e.g. "x1000"
//                nil on success, error info on error.

func CheckDomain(endpoints []string, domain string) (string,error) {
	var err error
	var domName string

	if (domain == CertDomainCabinet) {
		domName,err = checkDomainTargs(endpoints,domain,"c")
	} else if (domain == CertDomainChassis) {
		domName,err = checkDomainTargs(endpoints,domain,"s")
	} else if (domain == CertDomainBlade) {
		domName,err = checkDomainTargs(endpoints,domain,"b")
	} else if (domain == CertDomainBMC) {
		if (len(endpoints) > 1) {
			err = fmt.Errorf("BMC domain target list can only contain 1 target.")
		} else {
			domName,err = checkDomainTargs(endpoints,domain,"")
		}
	} else {
		err = fmt.Errorf("Invalid domain: '%s'",domain)
	}

	return domName,err
}

// Given a PEM encoded cert or key, convert all actual newline characters to
// a \n tuple.  If the input string already has \n tuples, do nothing.
//
// pemStr(in): PEM encoded cert or key string.
// Return:     Input data with newlines converted to tuples.

func NewlineToTuple(pemStr string) string {
	return strings.Replace(strings.Trim(pemStr,"\n"),"\n",`\n`,-1)
}

// Given a PEM encoded cert or key, convert all \n tuples to actual newline 
// characters.  If the input string already has newlines, do nothing.
//
// pemStr(in): PEM encoded cert or key string.
// Return:     Input data with tuples converted to newlines.

func TupleToNewline(pemStr string) string {
	return strings.Replace(pemStr,`\n`,"\n",-1)
}

// Create TLS cert/key pair for the specified endpoints.  The endpoints
// must be confined to the domain specified.  For example, if CertDomainCabinet
// is specified, all endpoints must reside in the same cabinet.
//
// If there is only one endpoint specified, then all possible components of 
// the specified type in the specified domain will be included in the key.
//
// Example, cert/key for sparse components:
//   endpoints: ["x0c0s0b0","x0c0s1b0","x0c0s2b0"], domain: cab
//      key will be for x0000 and have SANs for the endpoints listed.
//
// Example: cert/key for an entire cabinet:
//   endpoints: ["x1000"], domain: cab
//      key will be for x1000 and have SANs for all possible BMCs in the cab
//   
// endpoints(in): List of target BMCs.
// domain(in):    Target domain:
//                    CertDomainCabinet 
//                    CertDomainChassis 
//                    CertDomainBlade   
//                    CertDomainBMC   
// fqdn(in):      FQDN, e.g. "rocket.us.cray.com" to use in cert creation.
//                Can be empty.
// retData(out):  Returned TLS cert/key data.  Certs/keys are in JSON-frienly 
//                format.
// Return:        nil on succes, error string on error.

func CreateCert(endpoints []string, domain string, fqdn string,
                retData *VaultCertData) error {
	var vreq vaultCertReq

	domName, err := CheckDomain(endpoints, domain)
	if (err != nil) {
		return err
	}

	ss, err := sstorage.NewVaultAdapterAs(ConfigParams.VaultPKIBase, "pki-common-direct")
	if (err != nil) {
		return fmt.Errorf("ERROR creating secure storage adapter: %v", err)
	}

	//Create the request for vault certs

	vreq.CommonName = domName
	vreq.TTL = "8760h"	//1 year TODO: this may change.

	if (len(endpoints) == 1) {
		vreq.AltNames, err = genAllDomainAltNames(domName, domain)
		if (err != nil) {
			return err
		}
	} else {
		vreq.AltNames = strings.Join(endpoints, ",")
	}

	//Append FQDN to each AltName

	if (fqdn != "") {
		npfqdn := strings.TrimLeft(fqdn, ".")
		fqdn = "." + npfqdn
		anames := strings.Split(vreq.AltNames, ",")
		for ix := 0; ix < len(anames); ix ++ {
			anames[ix] = anames[ix] + fqdn
		}
		vreq.AltNames = strings.Join(anames, ",")
	}

	//Make the call to Vault
	err = ss.StoreWithData(ConfigParams.PKIPath, vreq, retData)
	if (err != nil) {
		return err
	}

	//Add in the FQDN for tracking
	retData.Data.FQDN = fqdn

	return nil
}

// Fetch the CA chain (a.k.a. 'bundle') cert.  
//
// uri(in): URI of CA chain data.  Can be a pathname or VaultCAChainURI
// Return:  CA bundle cert in JSON-friendly format.
//          nil on success, error string on error

func FetchCAChain(uri string) (string,error) {
	caChain := ""
	if (uri == VaultCAChainURI) {
		ss, err := sstorage.NewVaultAdapterAs(ConfigParams.VaultPKIBase, "pki-common-direct")
		if (err != nil) {
			return caChain, fmt.Errorf("ERROR creating secure storage adapter: %v", err)
		}
		
		err = ss.Lookup(ConfigParams.CAChainPath, &caChain)
		if (err != nil) {
			return caChain, fmt.Errorf("ERROR fetching CA Chain: %v", err)
		}
		return caChain, nil
	}

	//Nope, must be a file (from configmap)

	data,err := ioutil.ReadFile(uri)
	if (err != nil) {
		return "", fmt.Errorf("ERROR reading file '%s': %v", uri, err)
	}
	return string(data), nil
}

// Take a cert/key pair and store it in Vault.
//
// domainID(in):  Top-of-domain XName (e.g. x1000)
// certData(out): Returned cert info from Vault PKI.
// Return:        nil on success, error info on error.

func StoreCertData(domainID string, certData VaultCertData) error {
	if (!__vaultEnabled) {
		vstore[domainID] = certData
		return nil
	}
	ss,err := sstorage.NewVaultAdapter(ConfigParams.VaultKeyBase)
	if (err != nil) {
		return fmt.Errorf("ERROR creating secure storage adapter: %v",err)
	}

	err = ss.Store(vaultKey(domainID),&certData)
	if (err != nil) {
		return fmt.Errorf("ERROR storing to key '%s': %v",domainID,err)
	}

	return nil
}

// Delete a cert from Vault storage.
//
// domainID(in):  Cert domain ID (e.g. x1000 for a cabinet domain)
// force(in):     Non-existent cert is an error unless force=true.
// Return:        nil on success, error info on error.

func DeleteCertData(domainID string, force bool) error {
	if (!__vaultEnabled) {
		_,ok := vstore[domainID]
		if (!ok) {
			if (force) {
				logger.Infof("Can't find key: '%s', ignoring (force==true).",
					domainID)
				return nil
			} else {
				return fmt.Errorf("ERROR looking up '%s' (force==false).",
						domainID)
			}
		} else {
			delete(vstore,domainID)
		}
		return nil
	}
	ss,err := sstorage.NewVaultAdapter(ConfigParams.VaultKeyBase)
	if (err != nil) {
		return fmt.Errorf("ERROR creating secure storage adapter: %v",err)
	}

	//Since the Vault docs do everything in their power to obfuscate how
	//the interface behaves if the desired key doesn't exist, we'll take
	//safe route and read the key first to see if it exists.

	var cdata VaultCertData
	err = ss.Lookup(vaultKey(domainID),cdata)
	if (err != nil) {
		if (force) {
			logger.Infof("Can't find key: '%s', ignoring (force==true).",
				domainID)
			return nil
		} else {
			return fmt.Errorf("ERROR looking up '%s' (force==false): %v",
						domainID,err)
		}
	}

	//No error might mean there was nothing to read.  Check if the returned
	//data is empty to confirm.

	if ((cdata.RequestID == "") && (cdata.Data.Certificate != "")) {
		logger.Tracef("Target key '%s' not found in Vault, not deleting.",
				domainID)
		return nil
	}

	err = ss.Delete(vaultKey(domainID))
	if (err != nil) {
		return fmt.Errorf("ERROR deleting Vault key '%s': %v",
					domainID,err)
	}

	return nil
}

// Fetch a cert/key pair for a given XName within a given domain.
//
// xname(in):  Name of a BMC, OR, domain (e.g. x1000 for a cabinet domain)
// domain(in): BMC domain (e.g. hms_certs.CertDomainCabinet)
// Return:     Cert information for target;
//             nil on success, error info on error.

func FetchCertData(xname string, domain string) (VaultCertData,error) {
	var jdata VaultCertData

	domKey,derr := CheckDomain([]string{xname,},domain)
	if (derr != nil) {
		return jdata,fmt.Errorf("ERROR getting domain xname from '%s': %v",
					xname,derr)
	}

	if (!__vaultEnabled) {
		_,ok := vstore[domKey]
		if (ok) {
			return vstore[domKey],nil
		}
		return jdata,fmt.Errorf("ERROR fetching data for key '%s', xname '%s': Key does not exist.",
					domKey,xname)
	}

	ss,err := sstorage.NewVaultAdapter(ConfigParams.VaultKeyBase)
	if (err != nil) {
		return jdata,fmt.Errorf("ERROR creating secure storage adapter: %v",err)
	}

	//Create the storage key from the XName and domain

	err = ss.Lookup(vaultKey(domKey),&jdata)
	if (err != nil) {
		return jdata,fmt.Errorf("ERROR fetching data for key '%s', xname '%s': %v",
					domKey,xname,err)
	}

	return jdata,nil
}

// Do the grunt work of creating a secure HTTP client.  A retryable http client
// is used underneath, but it can be created with no retries to mimic a "plain"
// HTTP client.
//
// caURI(in):         CA trust bundle URI
// timeoutSecs(in)    Max timeout, in seconds.
// maxRetryCount(in): Max number of times to retry failures.  0 == try once.
// maxRetrySecs(in):  Max back-off time, in seconds, for retries.
// Return:            HTTP client, err string on error, nil on success.

func createSecHTTPClient(caURI string, timeoutSecs int,
                         maxRetryCount int, maxRetrySecs int) (*retryablehttp.Client,error) {
	caChain,err := FetchCAChain(caURI)
	if (err != nil) {
		return nil,err
	}

	certPool,cperr := x509.SystemCertPool()
	if (cperr != nil) {
		return nil,cperr
	}
	certPool.AppendCertsFromPEM([]byte(TupleToNewline(caChain)))
	tlsConfig := &tls.Config{RootCAs: certPool,}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig,}
	client := &http.Client{Transport: transport,
	                       Timeout: (time.Duration(timeoutSecs) * time.Second),}

	rtClient := retryablehttp.NewClient()
	rtClient.RetryMax = maxRetryCount
	rtClient.RetryWaitMax = time.Duration(maxRetrySecs) * time.Second
	rtClient.HTTPClient = client

	return rtClient,nil
}

// Given the URI (pathname or vault URI) of a CA cert chain bundle,
// create a secure "normal" (non-retrying) HTTP client.
//
// timeoutSecs(in): Timeout, in seconds, for HTTP transport/client connections
// caURI(in):       URI of CA chain data.  Can be a pathname or VaultCAChainURI
// Return:          Client for secure HTTP use.
//                  nil on success, non-nil error if something went wrong.

func CreateSecureHTTPClient(timeoutSecs int, caURI string) (*retryablehttp.Client,error) {
	cl,err := createSecHTTPClient(caURI,timeoutSecs,0,1)
	return cl,err
}

// Given the URI (pathname or vault URI) of a CA cert chain bundle,
// create a secure retryable (non-retrying) HTTP client.
//
// caURI(in):         CA trust bundle URI
// timeoutSecs(in)    Max timeout, in seconds.
// maxRetryCount(in): Max number of times to retry failures.  0 == try once.
// maxRetrySecs(in):  Max back-off time, in seconds, for retries.
// Return:            HTTP client pair, err string on error, nil on success.

func CreateRetryableSecureHTTPClient(caURI string, timeoutSecs int,
                                     maxRetryCount int, maxRetrySecs int) (*retryablehttp.Client,error) {
	cl,err := createSecHTTPClient(caURI,timeoutSecs,maxRetryCount,maxRetrySecs)
	return cl,err
}

// Do the grunt work of creating a non-TLS-validated HTTP client.
//
// timeoutSecs(in)    Max timeout, in seconds.
// maxRetryCount(in): Max number of times to retry failures.  0 == try once.
// maxRetrySecs(in):  Max back-off time, in seconds, for retries.
// Return:            HTTP client, err string on error, nil on success.

func createHTTPClient(timeoutSecs int, maxRetryCount int,
                      maxRetrySecs int) (*retryablehttp.Client,error) {
	client := &http.Client{Transport:
	              &http.Transport{TLSClientConfig:
	                  &tls.Config{InsecureSkipVerify: true,},},
	                  Timeout: (time.Duration(timeoutSecs) * time.Second),}
	rtClient := retryablehttp.NewClient()
	rtClient.RetryMax = maxRetryCount
	rtClient.RetryWaitMax = time.Duration(maxRetrySecs) * time.Second
	rtClient.HTTPClient = client
	return rtClient,nil
}

// Create a non-cert-verified HTTP transport, either normal or retryable.
//
// Args:   None.
// Return: Client for secure HTTP use.
//         nil on success, non-nil error if something went wrong.

func CreateInsecureHTTPClient(timeoutSecs int) (*retryablehttp.Client,error) {
	cl,err := createHTTPClient(timeoutSecs,0,1)
	return cl,err
}

func CreateRetryableInsecureHTTPClient(timeoutSecs int, maxRetryCount int,
                                       maxRetrySecs int) (*retryablehttp.Client,error) {
	cl,err := createHTTPClient(timeoutSecs,maxRetryCount,maxRetrySecs)
	return cl,err
}

// Create a struct containing both a cert-validated and a non-cert-validated
// HTTP client.
//
// caURI(in):       URI of CA chain data.  Can be a pathname or VaultCAChainURI
// timeoutSecs(in): Timeout, in seconds, for HTTP transport/client connections
// Return:          Client pair for secure and insecure HTTP use.
//                  nil on success, non-nil error if something went wrong.

func CreateHTTPClientPair(caURI string, timeoutSecs int) (*HTTPClientPair,error) {
	var secClient,insecClient *retryablehttp.Client
	var err error

	// No CA URI, create insecure transports and populate both with the same
	// transport.

	if (caURI == "") {
		logger.Warningf("CA URI is empty, creating non-cert validated HTTPS transport.")
		secClient,err = CreateInsecureHTTPClient(timeoutSecs)
		if (err != nil) {
			return nil,err
		}
		insecClient = secClient
	} else {
		secClient,err = CreateSecureHTTPClient(timeoutSecs,caURI)
		if (err != nil) {
			return nil,err
		}

		insecClient,err = CreateInsecureHTTPClient(timeoutSecs)
		if (err != nil) {
			return nil,err
		}
	}

	return &HTTPClientPair{SecureClient: secClient, InsecureClient: insecClient,}, nil
}

func CreateRetryableHTTPClientPair(caURI string, timeoutSecs int,
                                   maxRetryCount int, maxRetrySecs int) (*HTTPClientPair,error) {
	var secClient,insecClient *retryablehttp.Client
	var err error

	// No CA URI, create insecure transports and populate both with the same
	// transport.

	if (caURI == "") {
		logger.Warningf("CA URI is empty, creating non-cert validated HTTPS transport.")
		secClient,err = CreateRetryableInsecureHTTPClient(timeoutSecs,maxRetryCount,maxRetrySecs)
		if (err != nil) {
			return nil,err
		}
		insecClient = secClient
	} else {
		secClient,err = CreateRetryableSecureHTTPClient(caURI,timeoutSecs,maxRetryCount,maxRetrySecs)
		if (err != nil) {
			return nil,err
		}

		insecClient,err = CreateRetryableInsecureHTTPClient(timeoutSecs,maxRetryCount,maxRetrySecs)
		if (err != nil) {
			return nil,err
		}
	}

	return &HTTPClientPair{SecureClient: secClient, InsecureClient: insecClient,}, nil
}

func (p *HTTPClientPair) CloseIdleConnections() {
	if (p == nil) {
		return
	}
	p.FailedOver = false
	if (p.SecureClient != nil) {
		p.SecureClient.HTTPClient.CloseIdleConnections()
	}
	if (p.InsecureClient != nil) {
		p.InsecureClient.HTTPClient.CloseIdleConnections()
	}
}

func (p *HTTPClientPair) Do(req *http.Request) (*http.Response,error) {
	funcName := "HTTPClientPair.Do()"
	var rsp *http.Response
	var err error

	if (p == nil) {
		return rsp,fmt.Errorf("%s: Client pair is nil.",funcName)
	}
	rtReq,rtErr := retryablehttp.FromRequest(req)
	if (rtErr != nil) {
		return rsp,fmt.Errorf("%s: Can't create retryable HTTP request: %v",
					funcName,rtErr)
	}
	base.SetHTTPUserAgent(rtReq.Request,instName)
	p.FailedOver = false
	url := req.URL.Host + req.URL.Path

	if ((p.SecureClient == nil) && (p.InsecureClient == nil)) {
		return rsp,fmt.Errorf("%s: Client pair is uninitialized, not usable.",
					funcName)
	}

	if (p.SecureClient != nil) {
		rsp,err = p.SecureClient.Do(rtReq)
		if (err != nil) {
			// Do not attempt insecure if context was canceled as that means we
			// should abort and return.  Likely value in attempting insecure if
			// context dealine exceeded on secure attempt
			if p.InsecureClient != p.SecureClient && !errors.Is(err, context.Canceled) {

				seclog_Errorf("%s: TLS-secure transport failed for '%s': %v -- trying insecure client.",
						funcName,url,err)
				if (p.InsecureClient == nil) {
					emsg := fmt.Sprintf("%s: Failover to insecure transport failed: insecure client is nil.",funcName)
					seclog_Errorf("%s", emsg)
					return rsp,fmt.Errorf("%s", emsg)
				}

				p.FailedOver = true
				rsp,err = p.InsecureClient.Do(rtReq)
				if (err != nil) {
					seclog_Errorf("%s: TLS-insecure transport failed for '%s': %v",
						funcName,url,err)
					return rsp,err
				}
			} else {
				return rsp,err
			}
		}
	} else {
		seclog_Warnf("%s: TLS-secure transport not available, using insecure.",
				funcName)
		rsp,err = p.InsecureClient.Do(rtReq)
		if (err != nil) {
			seclog_Errorf("%s: TLS-insecure transport failed for '%s': %v",
				funcName,url,err)
			return rsp,err
		}
	}

	return rsp,nil
}

func (p *HTTPClientPair) Get(url string) (*http.Response,error) {
	funcName := "HTTPClientPair.Get()"
	var rsp *http.Response

	if (p == nil) {
		return rsp,fmt.Errorf("%s: Client pair is nil.",funcName)
	}

	req,_ := http.NewRequest("GET",url,nil)
	base.SetHTTPUserAgent(req,instName)
	return p.Do(req)
}

func (p *HTTPClientPair) Head(url string) (*http.Response,error) {
	funcName := "HTTPClientPair.Head()"
	var rsp *http.Response
	var err error

	if (p == nil) {
		return rsp,fmt.Errorf("%s: Client pair is nil.",funcName)
	}
	p.FailedOver = false

	if ((p.SecureClient == nil) && (p.InsecureClient == nil)) {
		return rsp,fmt.Errorf("%s: Client pair is uninitialized, not usable.",
					funcName)
	}

	if (p.SecureClient != nil) {
		rsp,err = p.SecureClient.Head(url)
		if (err != nil) {
			// Do not attempt insecure if context was canceled as that means we
			// should abort and return.  Likely value in attempting insecure if
			// context dealine exceeded on secure attempt
			if p.InsecureClient != p.SecureClient && !errors.Is(err, context.Canceled) {

				seclog_Errorf("%s: TLS-secure transport failed for '%s': %v -- trying insecure client.",
					funcName,url,err)
				if (p.InsecureClient == nil) {
					emsg := fmt.Sprintf("%s: Failover to insecure transport failed: insecure client is nil.",funcName)
					seclog_Errorf("%s", emsg)
					return rsp,fmt.Errorf("%s", emsg)
				}
				p.FailedOver = true
				rsp,err = p.InsecureClient.Head(url)
				if (err != nil) {
					seclog_Errorf("%s: TLS-insecure transport failed for '%s': %v",
						funcName,url,err)
					return rsp,err
				}
			} else {
				return rsp,err
			}
		}
	} else {
		seclog_Warnf("%s: TLS-secure transport not available, using insecure.",
			funcName)
		rsp,err = p.InsecureClient.Head(url)
		if (err != nil) {
			seclog_Errorf("%s: TLS-insecure transport failed for '%s': %v",
				funcName,url,err)
			return rsp,err
		}
	}

	return rsp,nil
}

func (p *HTTPClientPair) Post(url, contentType string, body io.Reader) (*http.Response,error) {
	funcName := "HTTPClientPair.Post()"
	var rsp *http.Response

	if (p == nil) {
		return rsp,fmt.Errorf("%s: Client pair is nil.",funcName)
	}

	req,_ := http.NewRequest("POST",url,body)
	req.Header.Add("Content-Type",contentType)
	base.SetHTTPUserAgent(req,instName)
	return p.Do(req)
}

func (p *HTTPClientPair) PostForm(url string, data url.Values) (*http.Response,error) {
	funcName := "HTTPClientPair.PostForm()"
	var rsp *http.Response

	if (p == nil) {
		return rsp,fmt.Errorf("%s: Client pair is nil.",funcName)
	}

	//Gotta emulate this, then call Do()
	vals := data.Encode()
	req,_ := http.NewRequest("POST",url,bytes.NewBuffer([]byte(vals)))
	base.SetHTTPUserAgent(req,instName)
	req.Header.Add("Content-Type","application/x-www-form-urlencoded")
	return p.Do(req)
}


//TODO: Need funcs:
// o SetTargetInfo(compList []HsmComponent, replace bool) error
//   Keep a map of which components will use TLS (HPE, Cray Mt.) and which won't
//   'replace' means make a new map, else just add to map
//   o /Inventory/HardwareByFRU Specify fruid, type, manufacturer? 
//   o Can use results to figure out which are Cray Mt and HPE iLO.
//
// o TargetHTTPClient(xname string) *http.Client
//   Given a client, return an HTTP client relevant (secure vs. non-secure)

