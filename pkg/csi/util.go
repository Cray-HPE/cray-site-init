/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package csi

// stringInSlice is shorthand
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
