/*
 MIT License

 (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP

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

package csm

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"

	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
)

// MinimumVersion defines the minimum CSI API version that this application is back-wards compatible
// with.
// NOTE: This should never change to a higher version, support should grow backwards or remain
//
//	the same (e.g. changing to v1.2 is acceptable, changing to v1.5 is not).
const MinimumVersion = "v1.4"

// APIKeyName defines the name of the parameter and input file's key that defines the target API version.
const APIKeyName = "csm-version"

// APIEnvName defines the environment variable to read the CSM installation version from for checking the API
// version against.
const APIEnvName = "CSM_RELEASE"

// normalize ensures that any given string is not a URL, does not have any suffix, has a leading "v".
func normalize(version string) string {
	exp := regexp.MustCompile(`^(v?)`)
	csm := regexp.MustCompile(`(?i)^v?(csm-?)`)
	tar := regexp.MustCompile(`(\.tar\.[bxg]z2?)$`)
	v := path.Base(version)
	v = csm.ReplaceAllString(
		v,
		"",
	)
	v = tar.ReplaceAllString(
		v,
		"",
	)
	v = exp.ReplaceAllString(
		v,
		"v",
	)
	return v
}

// currentVersion returns the currently targeted API version for the current set of inputs into CSI.
func currentVersion() (
	string, error,
) {
	v := viper.GetViper()
	version := v.GetString(APIKeyName)
	if version == "" {
		return "", fmt.Errorf(
			"unable to resolve %s from given parameters; key/parameter was undefined",
			APIKeyName,
		)
	}
	return normalize(version), nil
}

// IsCompatible returns whether the currently set version is compatible with this variant of CSI or not.
func IsCompatible() (
	string, error,
) {
	version, err := currentVersion()
	if err != nil {
		return version, err
	}
	min := semver.MajorMinor(semver.Canonical(normalize(MinimumVersion)))
	set := semver.MajorMinor(semver.Canonical(normalize(version)))
	if semver.Compare(
		set,
		min,
	) == -1 {
		return version, fmt.Errorf(
			"[%s < %s] this edition of `csi` is only compatible with CSM %s and higher but the given parameters have %s set to: %s",
			set,
			min,
			min,
			APIKeyName,
			version,
		)
	}
	return version, nil
}

// DetectedVersion returns the version of CSM detected in the environment by the APIEnvName.
func DetectedVersion() (
	string, error,
) {
	csmRelease, exists := os.LookupEnv(APIEnvName)
	if exists {
		csmVersion := normalize(csmRelease)
		log.Printf(
			"Interpreting [%s=%s] as %s",
			APIEnvName,
			csmRelease,
			csmVersion,
		)
		if !semver.IsValid(csmVersion) {
			return "", fmt.Errorf(
				"%s must be a semver version string r'v?[0-9]+\\.[0-9]+(\\.[0-9]+)?(-alpha|beta|rc\\.[0-9]+])?' but was %s",
				APIEnvName,
				csmRelease,
			)
		}
		return csmVersion, nil
	}
	// note this does not check Kubernetes, it is assumed to be running in an installation environment.
	return "", fmt.Errorf(
		"CSM installation or upgrade presence is unknown; %s was not found in the environment",
		APIEnvName,
	)
}

// Compare does a semver.Compare of the CurrentVersion against a given version.
func Compare(version string) (
	string, int,
) {

	currentVersion, err := currentVersion()
	if err != nil {
		return "", 1
	}

	c := semver.Canonical(normalize(currentVersion))
	v := semver.Canonical(normalize(version))

	cmp := semver.Compare(
		c,
		v,
	)
	return currentVersion, cmp
}

// CompareMajorMinor is the same as Compare but only considers the major and minor version numbers.
// Useful for resolving whether a given full-release equal to a parent release.
// (e.g. v1.2.3 vs v1.2.4 evaluates as equal).
func CompareMajorMinor(version string) (
	string, int,
) {

	currentVersion, err := currentVersion()
	if err != nil {
		return "", 1
	}

	c := semver.MajorMinor(semver.Canonical(normalize(currentVersion)))
	v := semver.MajorMinor(semver.Canonical(normalize(version)))

	cmp := semver.Compare(
		c,
		v,
	)
	return currentVersion, cmp
}
