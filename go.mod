module github.com/Cray-HPE/cray-site-init

replace k8s.io/client-go => k8s.io/client-go v0.19.4 // pinning this version for now.  Something is indirectly asking for an obsolete version

require (
	github.com/AlekSi/gocov-xml v0.0.0-20190121064608-3a14fb1c4737 // indirect
	github.com/Cray-HPE/hms-base v1.15.0
	github.com/Cray-HPE/hms-bss v1.9.5
	github.com/Cray-HPE/hms-s3 v1.9.2
	github.com/Cray-HPE/hms-shcd-parser v1.6.2
	github.com/Cray-HPE/hms-sls v1.10.4
	github.com/Cray-HPE/hms-smd v1.30.9
	github.com/asaskevich/govalidator v0.0.0-20200907205600-7a23bdc65eef
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/gocarina/gocsv v0.0.0-20200925213129-04be9ee2e1a2
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/tools v0.1.5 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
)

go 1.14
