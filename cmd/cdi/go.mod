module tags.cncf.io/container-device-interface/cmd/cdi

go 1.21

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/spf13/cobra v1.6.0
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v3 v3.0.5
	tags.cncf.io/container-device-interface v1.1.0
	tags.cncf.io/container-device-interface/schema v0.0.0
	tags.cncf.io/container-device-interface/specs-go v1.1.0
)

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	tags.cncf.io/container-device-interface => ../..
	tags.cncf.io/container-device-interface/schema => ../../schema
	tags.cncf.io/container-device-interface/specs-go => ../../specs-go
)
