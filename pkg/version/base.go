package version

var (
	sha1ver   string                   // sha1 revision used to build the program
	version   string                   // git tag version
	buildTime string                   // when the executable was built
	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
