module tags.cncf.io/container-device-interface

go 1.21

require (
	github.com/fsnotify/fsnotify v1.5.1
	github.com/opencontainers/runtime-spec v1.3.0
	github.com/stretchr/testify v1.12.1
	go.yaml.in/yaml/v3 v3.0.5
	golang.org/x/sys v0.19.0
	tags.cncf.io/container-device-interface/specs-go v1.1.0
)

replace tags.cncf.io/container-device-interface/specs-go => ./specs-go
