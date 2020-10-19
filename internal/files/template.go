/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"bytes"
	"text/template"
)

// WriteTemplate applies a config to a Template and writes the result to the path indicated
func WriteTemplate(path string, tpl template.Template, conf interface{}) error {
	var bs bytes.Buffer
	tpl.Execute(&bs, conf)
	writeFile(path, bs.String())
	return nil
}
