module tags.cncf.io/container-device-interface/schema

go 1.23

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v3 v3.0.5
	tags.cncf.io/container-device-interface v1.1.1
	tags.cncf.io/container-device-interface/specs-go v1.1.1
)

require (
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/opencontainers/runtime-spec v1.3.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	tags.cncf.io/container-device-interface => ../
	tags.cncf.io/container-device-interface/specs-go => ../specs-go
)
