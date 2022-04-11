package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Info is heavily informed by the Kubernetes Versioning System.
type Info struct {
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	FixVr      string `json:"fixvr"`
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

// String returns info as a human-friendly version string.
func (info Info) String() string {
	return info.GitVersion
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Info {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in ./base.go

	// If major and minor are not individually set with ldflags, we can
	// separate the .version file as passed with an ldflag
	if fsMajor == "" && fsMinor == "" {
		if fsVersion != "" {
			fsVersions := strings.Split(fsVersion, ".")
			fsMajor = fsVersions[0]
			fsMinor = fsVersions[1]
			fsFixVr = fsVersions[2]
		}
	}
	return Info{
		Major:      fsMajor,
		Minor:      fsMinor,
		FixVr:      fsFixVr,
		GitVersion: gitVersion,
		GitCommit:  sha1ver,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
