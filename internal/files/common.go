/*
Copyright 2020 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"bufio"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ImportConfig converts a configuration file to a viper
func ImportConfig(configfile string) (*viper.Viper, error) {
	dirname, filename := path.Split(configfile)
	extenstion := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, extenstion)

	config := viper.New()
	config.SetConfigType(strings.TrimPrefix(extenstion, "."))
	config.SetConfigName(name)
	config.AddConfigPath(dirname)
	err := config.ReadInConfig()
	if err != nil {
		return config, err
	}
	config.WatchConfig()
	return config, nil
}

// ExportConfig converts a viper to a file on disk
func ExportConfig(configfile string, config *viper.Viper) error {
	// TODO: Consider doing something of value here or simply
	// refactor it away
	return viper.WriteConfigAs(configfile)
}

// Generic and safe-ish file writing code
func writeFile(path string, contents string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	size, err := w.WriteString(contents)
	if err != nil {
		return err
	}
	w.Flush()
	log.Printf("wrote %d bytes to %s\n", size, path)
	return nil
}
