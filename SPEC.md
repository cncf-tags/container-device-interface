# Container Device Interface Specification

- [Version](#version)
- [Overview](#overview)
- [General considerations](#general-considerations)
- [CDI JSON Specification](#well-known-error-codes)
- [CDI CLI Specification](#well-known-error-codes)

## Version

This is CDI **spec** version **0.1.0**.

#### Released versions

Released versions of the spec are available as Git tags.

| Tag  | Spec Permalink   | Change |
| -----| -----------------| -------|

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

The key words "must", "must not", "required", "shall", "shall not", "should", "should not", "recommended", "may" and "optonal" are used as specified in [RFC 2119][rfc-2119].

[rfc-2119]: https://www.ietf.org/rfc/rfc2119.txt

## CDI JSON Specification

### JSON Definition

```
{
    "cdiVersion": "0.1.0",
    "kind": "<name>",
    "kindShort": ["<short-name>", ...],
    "containerRuntime": ["<container-runtime-name>"], (optional)

    "devices": [
        {
            "name": "<name>",
            "nameShort": ["<short-name>", "<short-name>"], (optional)

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
            "deviceNodes": [ (optional)
                {
                    "hostPath": "<path>",
                    "containerPath": "<path>",

                    // Cgroups permissions of the device, candidates are one or more of
                    // * r - allows container to read from the specified device.
                    // * w - allows container to write to the specified device.
                    // * m - allows container to create device files that do not yet exist.
                    "permissions": "<permissions>" (optional)
                }
            ]
            "mounts": [ (optional)
                {
                    "hostPath": "<source>",
                    "containerPath": "<destination>",
                    "options": "<OCI Mount Options>", (optional)
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
            ]
        }
    ]
}
```

#### Specification version

* `cdiVersion` (string, REQUIRED) MUST be in [Semantic Version 2.0](https://semver.org) and specifies the version of the CDI specification used by the vendor.

#### Kind

* `kind` (string, REQUIRED) field specifies a label which uniquely identifies the device vendor.
  It can be used to disambiguate the vendor that matches a device, e.g: `docker/podman run --device vendor.com/device=foo ...`.
    * The `kind` and `kindShort` labels have two segments: a prefix and a name, separated by a slash (/).
    * The name segment is required and must be 63 characters or less, beginning and ending with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (\_), dots (.), and alphanumerics between.
    * The prefix must be a DNS subdomain: a series of DNS labels separated by dots (.), not longer than 253 characters in total, followed by a slash (/).
    * Examples (not an exhaustive list):
      * Valid: `vendor.com/foo`, `foo.bar.baz/foo-bar123.B_az`.
      * Invalid: `foo`, `vendor.com/foo/`, `vendor.com/foo/bar`.

#### Container Runtime

* `containerRuntime` (array of string, OPTIONAL) indicates that this CDI specification targets only a specific container runtime.
  If this field is indicated, the container runtime MUST use the one that matches its name if non exists it should use the one that does not indicate any `containerRuntime` value.
  Possible values (not an exhaustive list): docker, podman, gvisor, lxc

#### CDI Devices

The `devices` field describes the set of hardware devices that can be requested by the container runtime user.
Note: For a CDI file to be valid, at least one entry must be specified in this array.

  * `devices` (array of objects, REQUIRED) list of devices provided by the vendor.
    * `name` (string, REQUIRED), name of the device, can be used to refer to it when requesting a device.
      * Beginning and ending with an alphanumeric character ([a-z0-9A-Z]) with dashes (-), underscores (\_), dots (.), and alphanumerics between.
      * e.g: `docker/podman run --device foo ...`
    * `nameShort` (array of strings, OPTIONAL), alternative names for the device. Can be used to reduce the CLI verbosity
      * Entries in the array MUST use the same schema as the entry for the `name` field
    * `containerEdits` (object, OPTIONAL) this field is described in the next section.
      * This field should only be merged in the OCI spec if the device has been requested by the container runtime user.


#### OCI Edits

The `containerEdits` field describes edits to be made to the OCI specification. Currently only three kinds of edits can be made to the OCI specification: `devices`, `mounts` and `hooks`.

The `containerEdits` field is referenced in two places in the specification:
  * At the device level, where the edits MUST only be made if the matching device is requested by the container runtime user.
  * At the container level, where the edits MUST be made if any of the of the device defined in the `devices` field are requested.


The `containerEdits` field has the following definition:
  * `deviceNodes` (array of objects, OPTIONAL) describes the device nodes that should be mounted:
    * `hostPath` (string, REQUIRED) path of the device on the host.
    * `containerPath` (string, REQUIRED) path of the device within the container.
    * `permissions` (string, OPTIONAL) Cgroups permissions of the device, candidates are one or more of:
      * r - allows container to read from the specified device.
      * w - allows container to write to the specified device.
      * m - allows container to create device files that do not yet exist.
  * `mounts` (array of objects, OPTIONAL) describes the mounts that should be mounted:
    * `hostPath` (string, REQUIRED) path of the device on the host.
    * `containerPath` (string, REQUIRED) path of the device within the container.
    * `readonly` (boolean, OPTIONAL) If set, the mount is read-only.
    * `propagation` (string, OPTIONAL) Requested propagation mode, candidates are one of:
      * private - No mount propagation ("private" in Linux terminology).
      * hostToContainer - Mounts get propagated from the host to the container ("rslave" in Linux terminology).
      * bidirectional - Mounts get propagated from the host to the container and from the container to the host ("rshared" in Linux terminology).
  * `hooks` (array of objects, OPTIONAL) describes the hooks that should be ran:
    * `hookName` is the name of the hook to invoke, if the runtime is OCI compliant it should be one of {createRuntime, createContainer, startContainer, poststart, poststop}.
      Runtimes are free to allow custom hooks but it is advised for vendors to create a specific JSON file targeting that runtime
    * `path` (string, REQUIRED) with similar semantics to IEEE Std 1003.1-2008 execv's path. This specification extends the IEEE standard in that path MUST be absolute.
    * `args` (array of strings, OPTIONAL) with the same semantics as IEEE Std 1003.1-2008 execv's argv.
    * `env` (array of strings, OPTIONAL) with the same semantics as IEEE Std 1003.1-2008's environ.
    * `timeout` (int, OPTIONAL) is the number of seconds before aborting the hook. If set, timeout MUST be greater than zero. If not set container runtime will wait for the hook to return.

## Error Handling
  * Two or more files with identical `kind` values and identical `containerRuntime` field.
    Container runtimes should surface this error when any device with that `kind` is requested.
  * Kind requested is not present in any CDI file.
    Container runtimes should surface an error when a non existent kind is requested.
  * Device (not device node) Requested does not exist.
    Container runtimes should surface this error when a non existent device is requested.
  * "Resource" does not exist (e.g: Mount, Hook, ...).
    Container runtimes should surface this error when a non existent "resource" is requested (e.g: at "run" time).
    This is because a resource does not need to exist when the spec is written but it needs to exist when the container is created.
  * Hook fails to execute.
    Container runtimes should surface an error when hooks fails to execute.
