/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"bytes"
	"log"
	"text/template"
)

// WriteTemplate applies a config to a Template and writes the result to the path indicated
func WriteTemplate(path string, tpl *template.Template, conf interface{}) error {
	log.Printf("In WriteTemplate.\n")
	var bs bytes.Buffer
	tpl.Execute(&bs, conf)
	log.Printf("calling writefile with %v, %v", path, bs.String())
	return writeFile(path, bs.String())
}
