# Container Device Interface Specification

- [Version](#version)
- [Overview](#overview)
- [General considerations](#general-considerations)
- [CDI JSON Specification](#cdi-json-specification)
- [Error Handling](#error-handling)

## Version

This is CDI **spec** version **0.8.0**.

### Update policy

Any modifications to the **spec** will result in at least a minor version bump. When releasing changes
that only affect the API also implemented in this repository, the patch version will be bumped.

*Note*: The **spec** is still under active development and there exists the possibility of breaking changes being
introduced with new versions.
### Released versions

Released versions of the spec are available as Git tags.

| Tag  | Spec Permalink   | Change |
| -----| -----------------| -------|
| v0.3.0 |   | Initial tagged release of Spec |
| v0.4.0 |   | Added `type` field to Mount specification |
| v0.5.0 |   | Add `HostPath` to `DeviceNodes` |
| v0.6.0 |   | Add `Annotations` field to `Spec` and `Device` specifications |
|        |   | Allow dots (`.`)  in name segment of `Kind` field. |
| v0.7.0 |   | Add `IntelRdt`field. |
|        |   | Add `AdditionalGIDs` to `ContainerEdits` |
| v0.8.0 |   | Remove .ToOCI() functions from specs-go package. |

*Note*: The initial release of a **spec** with version `v0.x.0` will be tagged as
`v0.x.0` with subsequent changes to the API applicable to this version tagged as `v0.x.y`.
## Overview

The _Container Device Interface_, or _CDI_ describes a mechanism for container runtimes to create containers which are able to interact with third party devices.

For third party devices, it is often the case that interacting with these devices require container runtimes to expose more than a device node. For example a third party device might require you to have kernel modules loaded, host libraries mounted or specific procfs path exposed/masked.

The _Container Device Interface_ describes a mechanism which allows third party vendors to perform these operations such that it doesn't require changing the container runtime.

The mechanism used is a JSON file (similar the [Container Network Interface (CNI)][cni]) which allows vendors to describe the operations the container runtime should perform on the container's [OCI specification][oci].

The Container Device Interface enables the following two flows:

A. Device Installation
   1. A user installs a third party device driver (and third party device) on a machine.
   2. The device driver installation software writes a JSON file at a well known path (`/etc/cdi/vendor.json`).

B. Container Runtime
   1. A user runs a container with the argument `--device` followed by a device name.
   2. The container runtime reads the JSON file.
   3. The container runtime validates that the device is described in the JSON file.
   4. The container runtime pulls the container image.
   5. The container runtime generates an OCI specification.
   6. The container runtime transforms the OCI specification according to the instructions in the JSON file


[cni]: https://github.com/containernetworking/cni
[oci]: https://github.com/opencontainers/runtime-spec

## General considerations

- The device configuration is in JSON format and can easily be stored in a file.
- The device configuration includes mandatory fields such as "name" and "version".
- Fields in CDI structures are required unless specifically marked optional.

For the purposes of this proposal, we define the following terms:
- _container_ is an instance of code running in an isolated execution environment specified by the OCI image and runtime specifications.
- _device_ which refers to an actual hardware device.
- _runtime engine_ is a component that creates a container from an OCI Runtime Specification and a rootfs directory
- _container runtime_ which refers to the higher level component users tend to interact with for managing containers. It may also include lower level components that implement management of containers and pods (sets of containers). e.g: docker, podman, ...
- _container runtime interface integration_ which refers to a server that implements the Container Runtime Interface (CRI) services, e.g: containerd+cri, cri-o, ...

The keywords "must", "must not", "required", "shall", "shall not", "should", "should not", "recommended", "may" and "optional" are used as specified in [RFC 2119][rfc-2119].

[rfc-2119]: https://www.ietf.org/rfc/rfc2119.txt

## CDI JSON Specification

### JSON Definition

```
{
    "cdiVersion": "0.7.0",
    "kind": "<name>",

    // This field contains a set of key-value pairs that may be used to provide
    // additional information to a consumer on the spec.
    "annotations": { (optional)
        "key": "value"
    },

    "devices": [
        {
            "name": "<name>",

            // This field contains a set of key-value pairs that may be used to provide
            // additional information to a consumer on the specific device.
            "annotations": { (optional)
              "key": "value"
            },

            // Same as the below containerSpec field.
            // This field should only be applied to the Container's OCI spec
            // if that specific device is requested.
            "containerEdits": { ... }
        }
    ],

    // This field should be applied to the Container's OCI spec if any of the
    // devices defined above are requested on the CLI
    "containerEdits": [
        {
            "env": [ (optional)
                "<envName>=<envValue>"
            ]
            "deviceNodes": [ (optional)
                {
                    "path": "<path>",
                    "hostPath": "<hostPath>" (optional),
                    "type": "<type>" (optional),
                    "major": <int32> (optional),
                    "minor": <int32> (optional),
                    // file mode for the device
                    "fileMode": <int> (optional),
                    // Cgroups permissions of the device, candidates are one or more of
                    // * r - allows container to read from the specified device.
                    // * w - allows container to write to the specified device.
                    // * m - allows container to create device files that do not yet exist.
                    "permissions": "<permissions>" (optional),
                    "uid": <int> (optional),
                    "gid": <int> (optional)
                }
            ]
            "mounts": [ (optional)
                {
                    "hostPath": "<source>",
                    "containerPath": "<destination>",
                    "type": "<OCI Mount Type>", (optional)
                    "options": "<OCI Mount Options>" (optional)
                }
            ],
            "hooks": [ (optional)
                {
                    "hookName": "<hookName>",
                    "path": "<path>",
                    "args": ["<arg>", "<arg>"], (optional)
                    "env":  [ "<envName>=<envValue>"], (optional)
                    "timeout": <int> (optional)
                }
            ],
            // Additional GIDs to add to the container process.
            // Note that a value of 0 is ignored.
            additionalGIDs: [ (optional)
              <uint32>
            ]
            "intelRdt": { (optional)
                "closID": "<name>", (optional)
                "l3CacheSchema": "string" (optional)
                "memBwSchema": "string" (optional)
                "enableCMT": "<boolean>" (optional)
                "enableMBM": "<boolean>" (optional)
            }
        }
    ]
}
```

#### Specification version

* `cdiVersion` (string, REQUIRED) MUST be in [Semantic Version 2.0](https://semver.org) and specifies the version of the CDI specification used by the vendor.

#### Kind

* `kind` (string, REQUIRED) field specifies a label which uniquely identifies the device vendor.
  It can be used to disambiguate the vendor that matches a device, e.g: `docker/podman run --device vendor.com/device=foo ...`.
    * The `kind` label has two segments: a prefix and a name, separated by a slash (/).
    * The name segment is required and must be 63 characters or less, beginning and ending with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (\_), dots (.), and alphanumerics between.
    * The prefix must be a DNS subdomain: a series of DNS labels separated by dots (.), not longer than 253 characters in total, followed by a slash (/).
    * Examples (not an exhaustive list):
      * Valid: `vendor.com/foo`, `foo.bar.baz/foo-bar123.B_az`.
      * Invalid: `foo`, `vendor.com/foo/`, `vendor.com/foo/bar`.

#### CDI Devices

The `devices` field describes the set of hardware devices that can be requested by the container runtime user.
Note: For a CDI file to be valid, at least one entry must be specified in this array.

  * `devices` (array of objects, REQUIRED) list of devices provided by the vendor.
    * `name` (string, REQUIRED), name of the device, can be used to refer to it when requesting a device.
      * Beginning and ending with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (\_), dots (.), and alphanumerics between.
      * e.g: `docker/podman run --device foo ...`
      * Entries in the array MUST use the same schema as the entry for the `name` field
    * `containerEdits` (object, OPTIONAL) this field is described in the next section.
      * This field should only be merged in the OCI spec if the device has been requested by the container runtime user.


#### OCI Edits

The `containerEdits` field describes edits to be made to the OCI specification. Currently, the following kinds of edits can be made to the OCI specification: `env`, `devices`, `mounts` and `hooks`.

The `containerEdits` field is referenced in two places in the specification:
  * At the device level, where the edits MUST only be made if the matching device is requested by the container runtime user.
  * At the container level, where the edits MUST be made if any of the device defined in the `devices` field are requested.


The `containerEdits` field has the following definition:
  * `env` (array of strings in the format of "VARNAME=VARVALUE", OPTIONAL) describes the environment variables that should be set. These values are appended to the container environment array.
  * `deviceNodes` (array of objects, OPTIONAL) describes the device nodes that should be mounted:
    * `path` (string, REQUIRED) path of the device within the container.
    * `hostPath` (string, OPTIONAL) path of the device node on the host. If not specified the value for `path` is used.
    * `type` (string, OPTIONAL) Device type: block, char, etc.
    * `major` (int64, OPTIONAL) Device major number.
    * `minor` (int64, OPTIONAL) Device minor number.
    * `fileMode` (int64, OPTIONAL) file mode for the device.
    * `permissions` (string, OPTIONAL) Cgroups permissions of the device, candidates are one or more of:
      * r - allows container to read from the specified device.
      * w - allows container to write to the specified device.
      * m - allows container to create device files that do not yet exist.
    * `uid` (uint32, OPTIONAL) id of device owner in the container namespace.
    * `gid` (uint32, OPTIONAL) id of device group in the container namespace.
  * `mounts` (array of objects, OPTIONAL) describes the mounts that should be mounted:
    * `hostPath` (string, REQUIRED) path of the device on the host.
    * `containerPath` (string, REQUIRED) path of the device within the container.
    * `type` (string, OPTIONAL) the type of the filesystem to be mounted. For bind mounts (when options include either bind or rbind), the type is a dummy, often "none" (not listed in /proc/filesystems).
    * `options` (array of strings, OPTIONAL) Mount options of the filesystem to be used.
  * `hooks` (array of objects, OPTIONAL) describes the hooks that should be ran:
    * `hookName` is the name of the hook to invoke, if the runtime is OCI compliant it should be one of {createRuntime, createContainer, startContainer, poststart, poststop}.
      Runtimes are free to allow custom hooks but it is advised for vendors to create a specific JSON file targeting that runtime
    * `path` (string, REQUIRED) with similar semantics to IEEE Std 1003.1-2008 execv's path. This specification extends the IEEE standard in that path MUST be absolute.
    * `args` (array of strings, OPTIONAL) with the same semantics as IEEE Std 1003.1-2008 execv's argv.
    * `env` (array of strings, OPTIONAL) with the same semantics as IEEE Std 1003.1-2008's environ.
    * `timeout` (int, OPTIONAL) is the number of seconds before aborting the hook. If set, timeout MUST be greater than zero. If not set container runtime will wait for the hook to return.
  * `intelRdt` (object, OPTIONAL) describes the Linux [resctrl][resctrl] settings for the container (object, OPTIONAL)
    * `closID` (string, OPTIONAL) name of the `CLOS` (Class of Service).
    * `l3CacheSchema` (string, OPTIONAL) L3 cache allocation schema for the `CLOS`.
    * `memBwSchema` (string, OPTIONAL) memory bandwidth allocation schema for the `CLOS`.
    * `enableCMT` (boolean, OPTIONAL) whether to enable cache monitoring
    * `enableMBM` (boolean, OPTIONAL) whether to enable memory bandwidth monitoring
  * `additionalGids` (array of uint32s, OPTIONAL) A list of additional group IDs to add with the container process. These values are added to the `user.additionalGids` field in the OCI runtime specification. Values of 0 are ignored.

## Error Handling
  * Kind requested is not present in any CDI file.
    Container runtimes should surface an error when a non-existent kind is requested.
  * Device (not device node) Requested does not exist.
    Container runtimes should surface this error when a non existent device is requested.
  * "Resource" does not exist (e.g: Mount, Hook, ...).
    Container runtimes should surface this error when a non-existent "resource" is requested (e.g: at "run" time).
    This is because a resource does not need to exist when the spec is written, but it needs to exist when the container is created.
  * Hook fails to execute.
    Container runtimes should surface an error when hooks fails to execute.

[resctrl]: https://docs.kernel.org/arch/x86/resctrl.html
