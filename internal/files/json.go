/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"encoding/json"
)

// WriteJSONConfig marshals from an interface to json and writes the result to the path indicated
func WriteJSONConfig(path string, conf interface{}) error {
	bs, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	return writeFile(path, string(bs))
}
