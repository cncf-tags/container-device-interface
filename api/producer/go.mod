module tags.cncf.io/container-device-interface/api/producer

go 1.20

require (
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.1.0
	sigs.k8s.io/yaml v1.3.0
	tags.cncf.io/container-device-interface/specs-go v0.8.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace tags.cncf.io/container-device-interface/specs-go => ../../specs-go
