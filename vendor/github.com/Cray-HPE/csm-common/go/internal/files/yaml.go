/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"io"

	"gopkg.in/yaml.v2"
)

// EncodeYAML encodes object to writer
func EncodeYAML(f io.Writer, v interface{}) error {
	return yaml.NewEncoder(f).Encode(v)
}

// DecodeYAML decodes object from reader
func DecodeYAML(f io.Reader, v interface{}) error {
	return yaml.NewDecoder(f).Decode(v)
}

// WriteYAMLConfig marshals from an interface to yaml and writes the result to the path indicated
func WriteYAMLConfig(path string, conf interface{}) error {
	return WriteConfig(EncodeYAML, path, conf)
}

// ReadYAMLConfig unmarshals a YAML encoded object from the specified file
func ReadYAMLConfig(path string, conf interface{}) error {
	return ReadConfig(DecodeYAML, path, conf)
}
