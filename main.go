/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/
package main

import (
	"math/rand"
	"time"

	"github.com/Cray-HPE/cray-site-init/cmd"
)

func main() {
	// We use randomness in several places and need to seed it properly
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
