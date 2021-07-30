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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// Error and debug logging.  For now just send to stdout.
var rfDebug int = 0
var rfVerbose int = 0
var errlog *log.Logger = log.New(os.Stdout, "", log.LstdFlags)

func SetDebug(level int) {
	rfDebug = level
}

func SetVerbose(level int) {
	rfVerbose = level
}

func SetLogger(l *log.Logger) {
	errlog = l
}

// If json.Unmarshal failed with a single-field error or something worse.
func IsUnmarshalTypeError(err error) bool {
	_, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		return false
	} else {
		return true
	}
}

// Given a valid MAC in 12-byte hex format (with/without ":"), return a
// MAC address that is the original value plus offset, which can be
// negative.
//
// Returns: New MAC string in ":" hex format (keeping upper/lowercase unless
//          case is mixed, then defaulting to lower) and err == nil
//       OR Empty-String/Non-NIL-error if MAC string is invalid/bad-format
//          or would over/underflow.
func GetOffsetMACString(mac string, offset int64) (string, error) {
	separator := ":"
	macStripped := strings.Replace(mac, ":", "", 5)
	if mac == macStripped {
		macStripped = strings.Replace(mac, "-", "", 5)
		if mac != macStripped {
			separator = "-"
		}
	}
	if len(macStripped) != 12 {
		err := fmt.Errorf("MAC string has invalid length")
		return "", err
	}
	val48, err := strconv.ParseInt(macStripped, 16, 64)
	if err != nil {
		return "", err
	}
	newVal48 := val48 + offset
	if newVal48 < 0 || newVal48 > 0xffffffffffff {
		err := fmt.Errorf("MAC string would overflow")
		return "", err
	}
	if strings.ToUpper(mac) == mac {
		macStripped = fmt.Sprintf("%012X", newVal48)
	} else {
		macStripped = fmt.Sprintf("%012x", newVal48)
	}
	finalVal := macStripped[0:2]
	for i := 2; i < 12; i += 2 {
		finalVal = finalVal + separator + macStripped[i:(i+2)]
	}
	return finalVal, err
}

// Normalize the MAC string to be lower case, ":"-separated 6-hex-byte format.
// If the MAC does not seem to be valid (strange separator, wrong length, etc.)
// return the EMPTY string
func NormalizeMACIfValid(mac string) string {
	normMAC, err := NormalizeVerifyMAC(mac)
	if err != nil {
		return ""
	}
	return normMAC
}

// Normalize the MAC string to be lower case, ":"-separated 6-hex-byte format.
// If the MAC does not seem to be valid (strange separator, wrong length, etc.)
// return the ORIGINAL string
func NormalizeMAC(mac string) string {
	normMAC, err := NormalizeVerifyMAC(mac)
	if err != nil {
		return mac
	}
	return normMAC
}

// Normalize the MAC string to be lower case, ":"-separated 6-hex-byte format.
// If the MAC does not seem to be valid (strange separator, wrong length, etc.)
// return the EMPTY string AND set err != nil
func NormalizeVerifyMAC(mac string) (string, error) {
	macLower := strings.ToLower(mac)
	macStripped := strings.Replace(macLower, "-", "", 5)
	macStripped = strings.Replace(macStripped, ":", "", 5)
	macStripped = strings.Replace(macStripped, ".", "", 5)
	macStripped = strings.TrimSpace(macStripped)
	if len(macStripped) != 12 {
		err := fmt.Errorf("MAC string has invalid length")
		return "", err
	}
	val48, err := strconv.ParseInt(macStripped, 16, 64)
	if err != nil {
		return "", err
	}
	macStripped = fmt.Sprintf("%012x", val48)
	finalVal := macStripped[0:2]
	for i := 2; i < 12; i += 2 {
		finalVal = finalVal + ":" + macStripped[i:(i+2)]
	}
	return finalVal, err
}

// Normalize and compare two MAC strings and report -1 if mac1 is lower,
// 1 if mac1 is higher, and 0 if they match.
func MACCompare(mac1, mac2 string) (int, error) {
	macStripped1 := strings.Replace(mac1, ":", "", 5)
	macStripped1 = strings.Replace(macStripped1, "-", "", 5)
	macStripped1 = strings.Replace(macStripped1, ".", "", 5)
	if len(macStripped1) != 12 {
		err := fmt.Errorf("MAC1 string has invalid length")
		return -1, err
	}
	macStripped2 := strings.Replace(mac2, ":", "", 5)
	macStripped2 = strings.Replace(macStripped2, "-", "", 5)
	macStripped2 = strings.Replace(macStripped2, ".", "", 5)
	if len(macStripped2) != 12 {
		err := fmt.Errorf("MAC2 string has invalid length")
		return -1, err
	}
	val1, err := strconv.ParseInt(macStripped1, 16, 64)
	if err != nil {
		return -1, err
	}
	val2, err := strconv.ParseInt(macStripped2, 16, 64)
	if err != nil {
		return -1, err
	}
	if val1 < val2 {
		return -1, nil
	} else if val1 > val2 {
		return 1, nil
	}
	// equal
	return 0, nil
}

//
// Auto-generate http mock representations for each endpoint for regression
// testing.
//

var genTestingPayloadsTitle = ""
var genTestingPayloadsDumpEpID = ""
var genTestingPayloadsOutfile *os.File = nil

// Turn on dumping of http output with path info, formatted for http mock
// responses.  format is ep_id:ep_title.  ep_id is the single
// endpoint to dump and ep_title(optional) is to be used in variable names to
// make them unique.  If ep_title is not included, the ep_id will be used.
//
// NOTE: Not tread safe.  Set once per process before invoking client commands.
func EnableGenTestingPayloads(ep_id_and_title string) error {
	alphaNumFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}
	// Up to two args divided by separator
	args := strings.SplitN(ep_id_and_title, ":", 3)
	if len(args) < 1 || len(args) > 2 {
		return fmt.Errorf("wrong number of args or failed to parse")
	}
	// Make sure id is alphanumeric.
	split := strings.FieldsFunc(args[0], alphaNumFunc)
	if len(split) != 1 {
		return fmt.Errorf("ep_id contains non-alphanumeric characters")
	}
	ep_id := args[0]
	ep_title := ep_id
	file_title := ep_id
	if len(args) == 2 {
		split := strings.FieldsFunc(args[1], alphaNumFunc)
		if len(split) == 1 {
			ep_title = args[1]
			file_title += "_" + ep_title
		} else {
			return fmt.Errorf("ep_title contains non-alphanumeric characters")
		}
	}
	// Create/truncate output file.
	var err error
	tmpFile := filepath.Join(os.TempDir(), file_title)
	genTestingPayloadsOutfile, err = os.Create(tmpFile)
	if err != nil {
		genTestingPayloadsOutfile = nil
		return err
	}
	// Set global variables now that there are no errors.
	genTestingPayloadsDumpEpID = ep_id
	genTestingPayloadsTitle = ep_title
	return nil
}

// This dumps a particular endpoint in such a way that it can be used as
// a mock remote endpoint in unit tests.  There should be one per type of
// endpoint (HMS type, manufacturer, base model, major firmware version/vendor
// etc.).  Whenever interesting differences present themselves, basically.
//
// Output is to a file.  Note that '>' is used to filter case statement entries
// so that they can be organized later as a single block (it will be
// interleaved due to execution order otherwise).
func GenTestingPayloads(f *os.File, name, path string, payload []byte) error {
	pathStripped := strings.Replace(path, ".", "", -1)
	pathStripped = strings.Replace(pathStripped, "-", "", -1)
	pathStripped = strings.Replace(pathStripped, "%", "_", -1)
	pathStripped = strings.Replace(pathStripped, "/redfish/v1/", "", 1)
	pathStripped = strings.Replace(pathStripped, "/", "_", -1)

	pathVarName := "testPath" + name + pathStripped
	payloadVarName := "testPayload" + name + pathStripped

	_, err := fmt.Fprintf(f, "const %s = \"%s\"\n\n", pathVarName, path)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(f, "const %s = `\n%s`\n\n", payloadVarName, payload)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f, ">\t\tcase \"https://\" + testFQDN + %s:\n"+
		">\t\t\treturn &http.Response{\n"+
		">\t\t\t\tStatusCode: 200,\n"+
		">\t\t\t\t// Send mock response for rpath\n"+
		">\t\t\t\tBody:       ioutil.NopCloser(bytes.NewBufferString(%s)),\n"+
		">\t\t\t\tHeader:     make(http.Header),\n"+
		">\t\t\t}\n\n",
		pathVarName, payloadVarName)
	if err != nil {
		return err
	}
	return nil
}

// Return the ip address string, in normalized form for ipv6, or the
// empty string if it is not a valid ipv4 or ipv6 address.
// Normalization for ipv6 involves lower-casing, removing leading zeros, and
// adding'[' ']' brackets, plus anything else net.ParseIP does.
// If a port is present, it is added after any required bracketing.
func GetIPAddressString(ip string) string {
	// Normalization - Lower case
	ipLower := strings.ToLower(ip)
	// strip off port, and any square brackets on ipv6 addresses with a port
	host, port, err := net.SplitHostPort(ipLower)
	if err != nil {
		// No port, so brackets not required, but also not stripped by call.
		host = strings.Trim(ipLower, "[]")
		port = ""
	} else {
		// brackets stripped from host.
		if port != "" {
			port = ":" + port
		}
	}
	isIPv6 := false
	// Now that port has been removed, safe to remove zone, if expected
	addr, zone := SplitAddrZone(host)
	// Does non-zone portion appear to be IPv6?
	if strings.IndexAny(addr, ":") >= 0 {
		isIPv6 = true
	} else if zone != "" {
		// If no, shouldn't have zone
		return ""
	}
	// Zone, port stripped - now safe to verify the IP address
	netip := net.ParseIP(addr)
	if netip == nil {
		// Not a valid ipv4 or v6 address.
		return ""
	}
	// Add back zone if non-empty
	addrString := netip.String()
	if zone != "" {
		addrString = addrString + "%" + zone
	}
	if isIPv6 == true {
		// Add brackets for ipv6, they were removed if present earlier.
		addrString = "[" + addrString + "]"
	}
	// Add back port if it exists, since bracketing has been done.
	addrString = addrString + port
	return addrString
}

// Split a string into two parts at the LAST instance of token.
// strings functions make this annoying.
func StringSplitLast(s string, token byte) (prefix, suffix string) {
	i := len(s)
	for i--; i >= 0; i-- {
		if s[i] == token {
			break
		}
	}
	if i > 0 {
		prefix, suffix = s[:i], s[i+1:]
	} else {
		prefix = s
	}
	return prefix, suffix
}

// Split off the portion of the address string following the last '%'
func SplitAddrZone(ipaddr string) (addr, zone string) {
	// Want last instance of zone separator '%'
	addr, zone = StringSplitLast(ipaddr, '%')
	return
}
