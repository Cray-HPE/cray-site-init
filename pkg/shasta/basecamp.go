/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `yaml:"local-hostname" json:"local-hostname"`       // should be xname
	InstanceID       string `yaml:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string `yaml:"region" json:"region"`                       // unused currently
	AvailabilityZone string `yaml:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string `yaml:"shasta-role" json:"shasta-role"`             // map to HSM role
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData MetaData               `yaml:"meta-data" json:"meta-data"`
	UserData map[string]interface{} `yaml:"user-data" json:"user-data"`
}
