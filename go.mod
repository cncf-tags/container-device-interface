module tags.cncf.io/container-device-interface

go 1.20

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.1.0
	github.com/opencontainers/runtime-tools v0.9.1-0.20221107090550-2e043c6bd626
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.19.0
	gopkg.in/yaml.v3 v3.0.1
	sigs.k8s.io/yaml v1.4.0
	tags.cncf.io/container-device-interface/specs-go v1.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/opencontainers/selinux v1.10.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/syndtr/gocapability v0.0.0-20200815063812-42c35b437635 // indirect
	golang.org/x/mod v0.19.0 // indirect
)

replace tags.cncf.io/container-device-interface/specs-go => ./specs-go
