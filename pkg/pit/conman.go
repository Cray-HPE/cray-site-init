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

package pit

import (
	"text/template"

	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
	"github.com/spf13/viper"
)

// ConmanConfigTemplate manages the Conman Configuration
var ConmanConfigTemplate = []byte(` 
SERVER keepalive=ON
SERVER logdir="/var/log/conman"
SERVER logfile="conman.log"
SERVER loopback=ON
SERVER pidfile="/var/run/conman/conman.pid"
SERVER resetcmd="powerman -0 %N; sleep 3; powerman -1 %N"
SERVER timestamp=1h
GLOBAL seropts="115200,8n1"
GLOBAL log="console.%N"
GLOBAL logopts="sanitize,timestamp"
{{range .}}
console name="{{.Hostname}}-mgmt"     dev="ipmi:{{.IP}}" ipmiopts="U:{{.User}},P:{{.Pass}},W:solpayloadsize"
{{- end}}
`)

// WriteConmanConfig provides conman configuration for the installer
func WriteConmanConfig(path string, ncns []csi.LogicalNCN) {
	type conmanLine struct {
		Hostname string
		User     string
		IP       string
		Pass     string
	}
	v := viper.GetViper()
	ncnBMCUser := v.GetString("bootstrap-ncn-bmc-user")
	ncnBMCPass := v.GetString("bootstrap-ncn-bmc-pass")

	var conmanNCNs []conmanLine

	for _, k := range ncns {
		conmanNCNs = append(conmanNCNs, conmanLine{
			Hostname: k.Hostname,
			User:     ncnBMCUser,
			Pass:     ncnBMCPass,
			IP:       k.BmcIP,
		})
	}

	tpl6, _ := template.New("conmanconfig").Parse(string(ConmanConfigTemplate))
	csiFiles.WriteTemplate(path, tpl6, conmanNCNs)
}
