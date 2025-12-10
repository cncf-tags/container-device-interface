module tags.cncf.io/container-device-interface/cmd/cdi

go 1.21

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/opencontainers/runtime-tools v0.9.1-0.20251114084447-edf4cb3d2116
	github.com/spf13/cobra v1.6.0
	gopkg.in/yaml.v3 v3.0.1
	sigs.k8s.io/yaml v1.4.0
	tags.cncf.io/container-device-interface v1.1.0
	tags.cncf.io/container-device-interface/schema v0.0.0
	tags.cncf.io/container-device-interface/specs-go v1.1.0
)

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
)

replace (
	tags.cncf.io/container-device-interface => ../..
	tags.cncf.io/container-device-interface/schema => ../../schema
	tags.cncf.io/container-device-interface/specs-go => ../../specs-go
)
