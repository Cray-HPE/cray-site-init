/*
Copyright 2021 Hewlett Packard Enterprise Development LP
*/

package files

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/spf13/viper"
)

// ExportConfig converts a viper to a file on disk
func ExportConfig(configfile string, config *viper.Viper) error {
	// TODO: Consider doing something of value here or simply
	// refactor it away
	return viper.WriteConfigAs(configfile)
}

type encoder func(io.Writer, interface{}) error
type decoder func(io.Reader, interface{}) error

// WriteConfig encodes an object to the specified file
func WriteConfig(enc encoder, path string, conf interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	enc(w, conf)
	w.Flush()
	info, _ := f.Stat()
	log.Printf("wrote %d bytes to %s\n", info.Size(), path)
	return nil
}

// ReadConfig decodes an object from the specified file
func ReadConfig(dec decoder, path string, conf interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return dec(f, conf)
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
