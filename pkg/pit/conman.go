/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package pit

import (
	"text/template"

	"github.com/spf13/viper"
	csiFiles "github.com/Cray-HPE/cray-site-init/internal/files"
	"github.com/Cray-HPE/cray-site-init/pkg/csi"
)

// ConmanConfigTemplate manages the Conman Configuration
var ConmanConfigTemplate = []byte(` 
SERVER keepalive=ON
SERVER logdir="/var/log/"
SERVER logfile="conman.log"
SERVER loopback=ON
SERVER pidfile="/var/run/conman.pid"
SERVER resetcmd="powerman -0 %N; sleep 3; powerman -1 %N"
SERVER timestamp=1h
GLOBAL seropts="115200,8n1"
GLOBAL log="conman/console.%N"
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
