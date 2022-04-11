module github.com/Cray-HPE/cray-site-init

replace (
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/kubectl => k8s.io/kubectl v0.22.2
)

require (
	github.com/Cray-HPE/csm-common/go v0.0.0-20220411164326-93391ce233dc
	github.com/Cray-HPE/hms-base v1.15.2-0.20210928201115-8d9f61f26219
	github.com/Cray-HPE/hms-bss v1.9.5
	github.com/Cray-HPE/hms-s3 v1.10.0
	github.com/Cray-HPE/hms-shcd-parser v1.6.2
	github.com/Cray-HPE/hms-sls v1.13.0
	github.com/Cray-HPE/hms-smd v1.30.9
	github.com/aws/aws-sdk-go v1.40.14
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/mitchellh/mapstructure v1.4.3
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
)

go 1.16
