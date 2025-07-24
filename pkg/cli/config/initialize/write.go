/*
 * MIT License
 *
 * (C) Copyright 2025 Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */

package initialize

import (
	"github.com/Cray-HPE/cray-site-init/pkg/cli"
	"github.com/Cray-HPE/cray-site-init/pkg/version"
	"github.com/spf13/viper"
)

// UniqueCIDRKeys are unique CIDRs, they should never overlap.
var UniqueCIDRKeys = []string{
	"can-cidr",
	"chn-cidr4",
	"chn-cidr6",
	"cmn-cidr4",
	"cmn-cidr6",
	"hmn-cidr",
	"hmn-mtn-cidr",
	"mtl-cidr",
	"nmn-cidr",
	"nmn-mtn-cidr",
}

// DeprecatedKeys is a list of every key that is deprecated in Cobra.
var DeprecatedKeys []string

// NoWriteKeys are keys that shouldn't be written to a config file.
var NoWriteKeys = []string{
	"config",
	"csm-api-url",
	"help",
	"input-dir",
	"k8s-namespace",
	"k8s-secret-name",
}

var Aliases []string

type TemplateData struct {
	Data      interface{}
	Timestamp string
	Version   string
}

// RegisterAlias a viper alias for a given key, and adds the alias to the Aliases slice.
func RegisterAlias(alias string, key string) {
	Aliases = append(
		Aliases,
		alias,
	)
	v := viper.GetViper()
	v.RegisterAlias(
		alias,
		key,
	)
}

/*
MakeTemplateData creates a TemplateData struct using the given interface. The returned struct has useful runtime
data such as the program version and its runtime timestamp. This information can be used in templates to identify
where they came from.
*/
func MakeTemplateData(data interface{}) TemplateData {
	return TemplateData{
		Data:      data,
		Timestamp: cli.RuntimeTimestamp,
		Version:   version.Get().String(),
	}
}

func WriteConfigAs(path string) (err error) {
	delConfig := viper.New()
	err = delConfig.MergeConfigMap(viper.AllSettings())
	if err != nil {
		return err
	}
	delConfigMap := delConfig.AllSettings()
	for _, key := range NoWriteKeys {
		delConfig.Set(
			key,
			struct{}{},
		)
		delete(
			delConfigMap,
			key,
		)
	}
	for _, key := range DeprecatedKeys {
		delConfig.Set(
			key,
			struct{}{},
		)
		delete(
			delConfigMap,
			key,
		)
	}

	for _, key := range Aliases {
		delConfig.Set(
			key,
			struct{}{},
		)
		delete(
			delConfigMap,
			key,
		)
	}
	finalConfig := viper.New()
	err = finalConfig.MergeConfigMap(delConfigMap)
	if err != nil {
		return err
	}
	err = finalConfig.WriteConfigAs(path)
	return err
}
