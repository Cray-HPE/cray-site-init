/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package shasta

// PhoneHome should not exist in data.json before installation has started
type PhoneHome struct {
	PublicKeyDSA     string `form:"pub_key_dsa" json:"pub_key_dsa" binding:"omitempty"`
	PublicKeyRSA     string `form:"pub_key_rsa" json:"pub_key_rsa" binding:"omitempty"`
	PublicKeyECDSA   string `form:"pub_key_ecdsa" json:"pub_key_ecdsa" binding:"omitempty"`
	PublicKeyED25519 string `form:"pub_key_ed25519" json:"pub_key_ed25519,omitempty"`
	InstanceID       string `form:"instance_id" json:"instance_id" binding:"omitempty"`
	Hostname         string `form:"hostname" json:"hostname" binding:"omitempty"`
	FQDN             string `form:"fdqn" json:"fdqn" binding:"omitempty"`
}

// MetaData is part of the cloud-init stucture and
// is only used for validating the required fields in the
// `CloudInit` struct below.
type MetaData struct {
	Hostname         string `form:"local-hostname" json:"local-hostname"`       // should be xname
	InstanceID       string `form:"instance-id" json:"instance-id"`             // should be unique for the life of the image
	Region           string `form:"region" json:"region"`                       // unused currently
	AvailabilityZone string `form:"availability-zone" json:"availability-zone"` // unused currently
	ShastaRole       string `form:"shasta-role" json:"shasta-role"`             // map to HSM role
}

// CloudInit is the main cloud-init struct. Leave the meta-data, user-data, and phone home
// info as generic interfaces as the user defines how much info exists in it.
type CloudInit struct {
	MetaData  map[string]interface{} `form:"meta-data" json:"meta-data"`
	UserData  map[string]interface{} `form:"user-data" json:"user-data"`
	PhoneHome PhoneHome              `form:"phone-home" json:"phone-home" binding:"omitempty"`
}
