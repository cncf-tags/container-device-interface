# CDI - The Container Device Interface

**NOTE:** The API for injecting CDI devices that existed at `container-device-interface/pkg` has been removed. Users of this API should migrate to the one at `container-device-interface/pkg/cdi` as this is actively maintained.

## What is CDI?

CDI (Container Device Interface), is a [specification](SPEC.md), for container-runtimes, to support third-party devices.

It introduces an abstract notion of a device as a resource. Such devices are
uniquely specified by a fully-qualified name that is constructed from a vendor
ID, a device class, and a name that is unique per vendor ID-device class pair.

```
vendor.com/class=unique_name
```

The combination of vendor ID and device class (`vendor.com/class` in the above
example) is referred to as the device kind.

CDI concerns itself only with enabling containers to be device aware. Areas like
resource management are explicitly left out of CDI (and are expected to be
handled by the orchestrator). Because of this focus, the CDI specification is
simple to implement and allows great flexibility for runtimes and orchestrators.

Note: The CDI model is based on the Container Networking Interface (CNI) model
and [specification](https://github.com/containernetworking/cni/blob/main/SPEC.md).

## Why is CDI needed?

On Linux, enabling a container to be device aware used to be as simple as
exposing a device node in that container. However, as devices and software grows
more complex, vendors want to perform more operations, such as:

- Exposing a device to a container can require exposing more than one device
  node, mounting files from the runtime namespace, or hiding procfs entries.
- Performing compatibility checks between the container and the device (e.g: Can
  this container run on this device?).
- Performing runtime-specific operations (e.g: VM vs Linux container-based
  runtimes).
- Performing device-specific operations (e.g: scrubbing the memory of a GPU or
  reconfiguring an FPGA).

In the absence of a standard for third-party devices, vendors often have to
write and maintain multiple plugins for different runtimes or even directly
contribute vendor-specific code in the runtime. Additionally, runtimes don't
uniformly expose a plugin system (or even expose a plugin system at all) leading
to duplication of the functionality in higher-level abstractions (such as
[Kubernetes device plugins](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/)).

## How does CDI work?

For CDI to work the following needs to be done:

- CDI file containing updates for the OCI spec in JSON or YAML format (with a
  `.json` or `.yaml` file extension, respectively) should be present in a CDI
  spec directory. The default directories are `/etc/cdi` and `/var/run/cdi`, but
  may depend on your runtime configuration
- One or more fully-qualified device names should be passed to the runtime
  either using command line parameters for CLI tools such as podman or Docker,
  or using container annotations or CRI fields for CRI-O and containerd
- The container runtime should be able to find the CDI file by the device name
  and update the container config using CDI file content.

## How to build and install CDI CLI?

### What is the CDI CLI?

The `cdi` command-line tool is a utility for inspecting and interacting with the CDI (Container Device Interface) cache.
It allows developers and system administrators to:

- **List CDI Spec files**: View all available CDI specification files in the configured directories
- **List vendors**: Display registered device vendors in the CDI cache
- **List device classes**: Show available device classes from CDI Specs
- **List devices**: Enumerate all CDI devices available in the system
- **Validate specs**: Verify CDI specification files against the JSON schema
- **Inject devices**: Inject CDI device configurations into OCI runtime specifications
- **Monitor cache**: Watch for changes in the CDI cache and Spec directories
- **Resolve devices**: Resolve fully-qualified device names to their configurations

The CLI tool is particularly useful for debugging CDI configurations, validating spec files, and testing device assignments before deploying them in production environments.

### Building the CDI CLI

To build the CDI command-line tool from source:

```bash
# Clone the repository (if not already done)
git clone https://github.com/cncf-tags/container-device-interface.git
cd container-device-interface

# Build the binary
make
```

This will compile the `cdi` binary and place it in the `bin/` directory along with other utilities like `validate`.

### Install the CDI CLI

After building, install the binary to your system:

```bash
# Install to /usr/local/bin (requires sudo)
sudo install -m 0755 bin/cdi /usr/local/bin/cdi

# Verify installation
cdi --help
```

### Basic Usage

Once installed, you can use the `cdi` command to interact with CDI devices:

```bash
# List all CDI Spec files
cdi specs

# List all available CDI devices
cdi devices

# List all vendors
cdi vendors

# List all device classes
cdi classes

# Validate CDI Spec files
cdi validate

# Monitor CDI cache for changes
cdi monitor
```

Use `cdi --help` or `cdi <command> --help` for detailed information about each command and its options.

## How to configure CDI?

### CRI-O configuration

In CRI-O CDI support is enabled by default. It is configured with the default
`/etc/cdi, /var/run/cdi` CDI directory locations. Therefore, you can start using
CDI simply by dropping CDI configuration files in either of those directories,
static configuration into `/etc/cdi` and dynamically updated one into
`/var/run/cdi`. If you are unsure of the configured directories you can run this
command to find them out:

```bash
$ crio config |& grep -B1 -A5 cdi_spec_dirs
```

### containerd configuration

To enable and configure CDI support in the [containerd
runtime](https://github.com/containerd/containerd) 2 configuration options
`enable_cdi` and `cdi_spec_dirs` should be set in the
`plugins."io.containerd.grpc.v1.cri` section of the containerd configuration
file (`/etc/containerd/config.toml` by default):

```
[plugins."io.containerd.grpc.v1.cri"]
  enable_cdi = true
  cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
```

Remember to restart containerd for any configuration changes to take effect.

### Docker and Podman Configuration

Both [Docker Daemon](https://github.com/moby/moby) and
[podman](https://github.com/containers/podman) support CDI and process specified
`--device` flags directly. If fully-qualified device selectors
(e.g., `vendor.com/device=myDevice`) are included, the CDI specifications at the
default location (`/etc/cdi` and `/var/run/cdi`) are checked for matching
devices.

Podman does not require any specific configuration to enable CDI support.

Docker has CDI enabled by default beginning with version **28.2.0**.

#### Docker older than 28.2.0

Docker supports CDI since version **25.0.0**.

For Docker versions between **25.0.0** and **28.1.1**, you'll need to enable the CDI feature by including the following in the daemon configuration file (`/etc/docker/daemon.json` by default):

```json
{
  "features": {
    "cdi": true
  }
}
```
Remember to restart the Docker daemon for any configuration changes to take effect.

#### Podman

Although initial support was added in [`v3.2.0`](https://github.com/containers/podman/releases/tag/v3.2.0), this was updated for the tagged `v0.3.0` CDI spec in [`v4.1.0-rc.1`](https://github.com/containers/podman/releases/tag/v4.1.0-rc1) with [commit a234e4e](https://github.com/containers/podman/commit/a234e4e19662e172472877ce69523f4afea5c12e).

## Examples

### Full-blown CDI specification

```bash
$ mkdir /etc/cdi
$ cat > /etc/cdi/vendor.json <<EOF
{
  "cdiVersion": "0.6.0",
  "kind": "vendor.com/device",
  "devices": [
    {
      "name": "myDevice",
      "containerEdits": {
        "deviceNodes": [
          {"hostPath": "/vendor/dev/card1", "path": "/dev/card1", "type": "c", "major": 25, "minor": 25, "fileMode": 384, "permissions": "rw", "uid": 1000, "gid": 1000},
          {"path": "/dev/card-render1", "type": "c", "major": 25, "minor": 25, "fileMode": 384, "permissions": "rwm", "uid": 1000, "gid": 1000}
        ]
      }
    }
  ],
  "containerEdits": {
    "env": [
      "FOO=VALID_SPEC",
      "BAR=BARVALUE1"
    ],
    "deviceNodes": [
      {"path": "/dev/vendorctl", "type": "b", "major": 25, "minor": 25, "fileMode": 384, "permissions": "rw", "uid": 1000, "gid": 1000}
    ],
    "mounts": [
      {"hostPath": "/bin/vendorBin", "containerPath": "/bin/vendorBin"},
      {"hostPath": "/usr/lib/libVendor.so.0", "containerPath": "/usr/lib/libVendor.so.0"},
      {"hostPath": "tmpfs", "containerPath": "/tmp/data", "type": "tmpfs", "options": ["nosuid","strictatime","mode=755","size=65536k"]}
    ],
    "hooks": [
      {"hookName": "createContainer", "path": "/bin/vendor-hook" },
      {"hookName": "startContainer", "path": "/usr/bin/ldconfig" }
    ]
  }
}
EOF
```

Assuming this specification has been generated and is available in either
`/etc/cdi` or `/var/run/cdi` (or wherever a CDI-enabled consumer is configured
to read CDI specifications from), the devices can be accessed through their
fully-qualified device names.

For example, in the case of `podman` the CLI for accessing the device would be:
```
$ podman run --device vendor.com/device=myDevice ...
```

### Using Annotations per device to add meta-information

```bash
$ mkdir /etc/cdi
$ cat > /etc/cdi/vendor-annotations.json <<EOF
{
  "cdiVersion": "0.6.0",
  "kind": "vendor.com/device",
  "devices": [
    {
      "name": "myDevice",
      "annotations": {
        "whatever": "false"
        "whenever": "true"
      }
      "containerEdits": {
        "deviceNodes": [
          {"path": "/dev/vfio/71"}
        ]
      }
    }
  ]
}
EOF
```

### Tutorial

For a step-by-step guide on creating and using CDI specifications, see the [tutorial](TUTORIAL.md).

## Issues and Contributing

[Check out the Contributing document!](CONTRIBUTING.md)

* Please let us know by [filing a new issue](https://github.com/cncf-tags/container-device-interface/issues/new)
* You can contribute by opening a [pull request](https://help.github.com/articles/using-pull-requests/)
