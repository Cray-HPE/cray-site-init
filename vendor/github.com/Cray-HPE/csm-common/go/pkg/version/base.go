package version

var (
	sha1ver    string                   // sha1 revision used to build the program
	gitVersion string                   // git tag version
	fsVersion  string                   // .version contents
	buildTime  string                   // when the executable was built
	fsMajor    string                   // major version, always numeric
	fsMinor    string                   // minor version, always numeric
	fsFixVr    string                   // fixvr version (bugfix / hotfix / non-contextually changing mod, numeric possibly followed by "+"
	buildDate  = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
