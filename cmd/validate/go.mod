module tags.cncf.io/container-device-interface/cmd/validate

go 1.21

require tags.cncf.io/container-device-interface/schema v0.0.0

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/text v0.14.0 // indirect
	tags.cncf.io/container-device-interface v1.1.0 // indirect
	tags.cncf.io/container-device-interface/specs-go v1.1.0 // indirect
)

replace (
	tags.cncf.io/container-device-interface => ../..
	tags.cncf.io/container-device-interface/schema => ../../schema
	tags.cncf.io/container-device-interface/specs-go => ../../specs-go
)
