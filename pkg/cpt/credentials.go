/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package cpt

//PasswordCredential is a struct for holding username/password credentials
type PasswordCredential struct {
	Username   string `form:"username" json:"username"`
	Password   string `form:"password" json:"password"`
	ServiceURL string `form:"service_url" json:"service_url" binding:"omitempty"`
}
