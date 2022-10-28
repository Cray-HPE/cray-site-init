/*
 MIT License

 (C) Copyright 2022 Hewlett Packard Enterprise Development LP

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
