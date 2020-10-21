/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import "gopkg.in/yaml.v2"

// WriteYamlConfig marshals from an interface to yaml and writes the result to the path indicated
func WriteYamlConfig(path string, conf interface{}) error {
	bs, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return writeFile(path, string(bs))
}
