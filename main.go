/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/
package main

import (
	"math/rand"
	"time"

	"stash.us.cray.com/MTL/sic/cmd"
)

func main() {
	// We use randomness in several places and need to seed it properly
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
