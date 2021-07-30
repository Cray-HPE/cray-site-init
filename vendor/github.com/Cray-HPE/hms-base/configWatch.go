// MIT License
//
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package base

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

var configWatch *fsnotify.Watcher

type configIn struct {
	Defs struct {
		Role    []string `json:"Role"`
		SubRole []string `json:"SubRole"`
	} `json:"HMSExtendedDefinitions"`
}

// main
func InitTypes(configpath string) error {
	if configpath == "" {
		return fmt.Errorf("InitTypes: No config file to watch")
	}
	err := watchConfig(configpath)
	if err != nil {
		return err
	}
	return nil
}

func watchConfig(configpath string) error {
	if configWatch != nil {
		// Already watching a file
		return nil
	}
	// configpath must not be a directory
	if strings.HasSuffix(configpath, "/") {
		return fmt.Errorf("watchConfig: Configpath must not be a directory: %s", configpath)
	}
	// Must be the absolute path
	if !path.IsAbs(configpath) {
		return fmt.Errorf("watchConfig: Must be the absolute path: %s", configpath)
	}

	// Get the directory path
	configdir := path.Dir(configpath)
	// Load the file contents
	loadFile(configpath)

	// Create the file watcher
	configWatch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	done := make(chan bool)
	go func() {
		defer configWatch.Close()
		for {
			select {
			// watch for events
			case <-configWatch.Events:
				loadFile(configpath)
			// watch for errors
			case err := <-configWatch.Errors:
				log.Printf("ERROR: watchConfig: %s\n", err)

			case <-done:
				configWatch = nil
				return
			}
		}
	}()
	// Mounting a configmap makes the configfile a symlink which will
	// not trigger change events. Watch the directory instead for changes.
	if err := configWatch.Add(configdir); err != nil {
		done <- true
		return err
	}
	return nil
}

func loadFile(file string) {
	// Attempt to read from file
	config := new(configIn)

	f, err := os.Open(file)
	if err != nil {
		log.Printf("Warning: loadFile: Failed to open config %s: %s\n", file, err)
		return
	}
	defer f.Close()

	bytes, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(bytes, config)
	if err != nil {
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			log.Printf("Warning: loadFile: bad field(s) skipped: %s: %s\n", file, err)
		} else {
			log.Printf("Warning: loadFile: Failed to decode config %s: %s\n", file, err)
			return
		}
	}

	// Reload the role maps with the defaults + our extended values
	hmsRoleMap = map[string]string{}
	for key, val := range defaultHMSRoleMap {
		hmsRoleMap[key] = val
	}
	if config.Defs.Role != nil && len(config.Defs.Role) != 0 {
		for _, val := range config.Defs.Role {
			key := strings.ToLower(val)
			hmsRoleMap[key] = val
		}
	}
	hmsSubRoleMap = map[string]string{}
	for key, val := range defaultHMSSubRoleMap {
		hmsSubRoleMap[key] = val
	}
	if config.Defs.SubRole != nil && len(config.Defs.SubRole) != 0 {
		for _, val := range config.Defs.SubRole {
			key := strings.ToLower(val)
			hmsSubRoleMap[key] = val
		}
	}
}
