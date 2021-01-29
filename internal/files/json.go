/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"encoding/json"
	"io"
)

// EncodeJSON encodes object to writer
func EncodeJSON(f io.Writer, v interface{}) error {
	return json.NewEncoder(f).Encode(v)
}

// DecodeJSON decodes object from reader
func DecodeJSON(f io.Reader, v interface{}) error {
	return json.NewDecoder(f).Decode(v)
}

// WriteJSONConfig marshals from an interface to json and writes the result to the path indicated
func WriteJSONConfig(path string, conf interface{}) error {
	return WriteConfig(EncodeJSON, path, conf)
}

// ReadJSONConfig unmarshals a JSON encoded object from the specified file
func ReadJSONConfig(path string, conf interface{}) error {
	return ReadConfig(DecodeJSON, path, conf)
}
