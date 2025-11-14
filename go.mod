module tags.cncf.io/container-device-interface

go 1.21

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/opencontainers/runtime-tools v0.9.1-0.20251114084447-edf4cb3d2116
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.26.0
	gopkg.in/yaml.v3 v3.0.1
	sigs.k8s.io/yaml v1.4.0
	tags.cncf.io/container-device-interface/specs-go v1.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/moby/sys/capability v0.4.0 // indirect
	github.com/opencontainers/selinux v1.13.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
)

replace tags.cncf.io/container-device-interface/specs-go => ./specs-go
