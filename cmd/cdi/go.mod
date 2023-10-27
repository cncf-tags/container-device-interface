module github.com/container-orchestrated-devices/container-device-interface/cmd/cdi

go 1.19

require (
	github.com/container-orchestrated-devices/container-device-interface v0.0.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.1.0
	github.com/opencontainers/runtime-tools v0.9.1-0.20221107090550-2e043c6bd626
	github.com/spf13/cobra v1.6.0
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/opencontainers/selinux v1.10.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/container-orchestrated-devices/container-device-interface => ../..
