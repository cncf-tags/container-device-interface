package producer

import (
	"fmt"
	"strings"

	"tags.cncf.io/container-device-interface/pkg/parser"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

// GenerateSpecName generates a vendor+class scoped Spec file name. The
// name can be passed to WriteSpec() to write a Spec file to the file
// system.
//
// vendor and class should match the vendor and class of the CDI Spec.
// The file name is generated without a ".json" or ".yaml" extension.
// The caller can append the desired extension to choose a particular
// encoding. Otherwise WriteSpec() will use its default encoding.
//
// This function always returns the same name for the same vendor/class
// combination. Therefore it cannot be used as such to generate multiple
// Spec file names for a single vendor and class.
func GenerateSpecName(vendor, class string) string {
	return vendor + "-" + class
}

// GenerateTransientSpecName generates a vendor+class scoped transient
// Spec file name. The name can be passed to WriteSpec() to write a Spec
// file to the file system.
//
// Transient Specs are those whose lifecycle is tied to that of some
// external entity, for instance a container. vendor and class should
// match the vendor and class of the CDI Spec. transientID should be
// unique among all CDI users on the same host that might generate
// transient Spec files using the same vendor/class combination. If
// the external entity to which the lifecycle of the transient Spec
// is tied to has a unique ID of its own, then this is usually a
// good choice for transientID.
//
// The file name is generated without a ".json" or ".yaml" extension.
// The caller can append the desired extension to choose a particular
// encoding. Otherwise WriteSpec() will use its default encoding.
func GenerateTransientSpecName(vendor, class, transientID string) string {
	transientID = strings.ReplaceAll(transientID, "/", "_")
	return GenerateSpecName(vendor, class) + "_" + transientID
}

// GenerateNameForSpec generates a name for the given Spec using
// GenerateSpecName with the vendor and class taken from the Spec.
// On success it returns the generated name and a nil error. If
// the Spec does not contain a valid vendor or class, it returns
// an empty name and a non-nil error.
func GenerateNameForSpec(raw *cdi.Spec) (string, error) {
	vendor, class := parser.ParseQualifier(raw.Kind)
	if vendor == "" {
		return "", fmt.Errorf("invalid vendor/class %q in Spec", raw.Kind)
	}

	return GenerateSpecName(vendor, class), nil
}

// GenerateNameForTransientSpec generates a name for the given transient
// Spec using GenerateTransientSpecName with the vendor and class taken
// from the Spec. On success it returns the generated name and a nil error.
// If the Spec does not contain a valid vendor or class, it returns an
// an empty name and a non-nil error.
func GenerateNameForTransientSpec(raw *cdi.Spec, transientID string) (string, error) {
	vendor, class := parser.ParseQualifier(raw.Kind)
	if vendor == "" {
		return "", fmt.Errorf("invalid vendor/class %q in Spec", raw.Kind)
	}

	return GenerateTransientSpecName(vendor, class, transientID), nil
}
