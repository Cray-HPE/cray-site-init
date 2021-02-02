/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"bytes"
	"log"
	"text/template"
)

// WriteTemplate applies a config to a Template and writes the result to the path indicated
func WriteTemplate(path string, tpl *template.Template, conf interface{}) error {
	var bs bytes.Buffer
	err := tpl.Execute(&bs, conf)
	if err != nil {
		log.Printf("The error executing the template is %v \n", err)
	}
	// log.Printf("calling writefile with %v, %v", path, bs.String())
	return writeFile(path, bs.String())
}
