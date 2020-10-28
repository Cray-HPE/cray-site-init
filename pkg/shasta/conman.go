/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

// ConmanConfigTemplate manages the CAN portion of the DNSMasq configuration
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
console name="{{.Hostname}}-mgmt"     dev="ipmi:{{.BMCIP}}" ipmiopts="U:root,P:initial0,W:solpayloadsize"
{{- end}}
`)
