module stash.us.cray.com/MTL/sic

go 1.15

replace (
	stash.us.cray.com/HMS/hms-base => ../../hms/hms-base
	stash.us.cray.com/HMS/hms-compcredentials => ../../hms/hms-compcredentials
	stash.us.cray.com/HMS/hms-s3 => ../../hms/hms-s3
	stash.us.cray.com/HMS/hms-securestorage => ../../hms/hms-securestorage
	stash.us.cray.com/HMS/hms-shcd-parser => ../../hms/hms-shcd-parser
	stash.us.cray.com/HMS/hms-sls => ../../hms/hms-sls
)

require (
	github.com/gocarina/gocsv v0.0.0-20200827134620-49f5c3fa2b3e
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.0.0
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.7.1
	gopkg.in/yaml.v2 v2.2.4
	stash.us.cray.com/HMS/hms-sls v0.0.0-00010101000000-000000000000
)
