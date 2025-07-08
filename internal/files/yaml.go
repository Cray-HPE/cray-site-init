/*
 MIT License

 (C) Copyright 2022-2025 Hewlett Packard Enterprise Development LP

 Permission is hereby granted, free of charge, to any person obtaining a
 copy of this software and associated documentation files (the "Software"),
 to deal in the Software without restriction, including without limitation
 the rights to use, copy, modify, merge, publish, distribute, sublicense,
 and/or sell copies of the Software, and to permit persons to whom the
 Software is furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included
 in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 OTHER DEALINGS IN THE SOFTWARE.
*/

package files

import (
	"io"

	"gopkg.in/yaml.v3"
)

// EncodeYAML encodes object to writer
func EncodeYAML(f io.Writer, v interface{}) error {
	newEncoder := yaml.NewEncoder(f)
	newEncoder.SetIndent(2)
	return newEncoder.Encode(v)
}

// DecodeYAML decodes object from reader
func DecodeYAML(f io.Reader, v interface{}) error {
	return yaml.NewDecoder(f).Decode(v)
}

// WriteYAMLConfig marshals from an interface to yaml and writes the result to the path indicated
func WriteYAMLConfig(path string, conf interface{}) error {
	return WriteConfig(
		EncodeYAML,
		path,
		conf,
	)
}

// ReadYAMLConfig unmarshals a YAML encoded object from the specified file
func ReadYAMLConfig(path string, conf interface{}) error {
	return ReadConfig(
		DecodeYAML,
		path,
		conf,
	)
}
