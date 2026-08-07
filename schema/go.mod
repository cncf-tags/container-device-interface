module tags.cncf.io/container-device-interface/schema

go 1.21

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/stretchr/testify v1.7.0
	sigs.k8s.io/yaml v1.4.0
	tags.cncf.io/container-device-interface v1.1.0
	tags.cncf.io/container-device-interface/specs-go v1.1.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/opencontainers/runtime-spec v1.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	tags.cncf.io/container-device-interface => ../
	tags.cncf.io/container-device-interface/specs-go => ../specs-go
)
