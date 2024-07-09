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

- CDI file containing updates for the OCI spec in JSON format should be present
  in the CDI spec directory. Default directories are `/etc/cdi` and
  `/var/run/cdi`
- Fully qualified device name should be passed to the runtime either using
  command line parameters for podman or using container annotations for CRI-O
  and containerd
- Container runtime should be able to find the CDI file by the device name and
  update the container config using CDI file content.

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

### Docker configuration

To enable and configure CDI support in the [Docker Daemon](https://github.com/moby/moby)
[Docker 25](https://github.com/moby/moby/releases/tag/v25.0.0) or later is required.

In addition, the CDI feature must be enabled. That means including the following in the daemon
config (`/etc/docker/daemon.json` by default):

```json
{
  "features": {
    "cdi": true
  }
}
```

Remember to restart to Docker daemon for any configuration changes to take effect.

### Podman configuration

[podman](https://github.com/containers/podman) does not require any specific
configuration to enable CDI support and processes specified `--device` flags
directly. If fully-qualified device selectors (e.g.
`vendor.com/device=myDevice`) are included the CDI specifications at the default
location (`/etc/cdi` and `/var/run/cdi`) are checked for matching devices.

*Note:* Although initial support was added in
[`v3.2.0`](https://github.com/containers/podman/releases/tag/v3.2.0) this was
updated for the tagged `v0.3.0` CDI spec in
[`v4.1.0-rc.1`](https://github.com/containers/podman/releases/tag/v4.1.0-rc1)
with [commit
a234e4e](https://github.com/containers/podman/commit/a234e4e19662e172472877ce69523f4afea5c12e).

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
      {"createContainer": {"path": "/bin/vendor-hook"} },
      {"startContainer": {"path": "/usr/bin/ldconfig"} }
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


## Presentations on CDI

See the following presentations for more information on CDI:

* Kubecon EU 2023 - Node Resource Management: The Big Picture - https://sched.co/1HyVB - https://youtu.be/LjhgklNNAtA
* Kubecon NA 2023 – The hidden heroes behind AI: Making sense of GPUs and TPUs in k8s - https://sched.co/1R2sf - https://youtu.be/H5gN8fHn3fo
* Kubecon EU 2024 – Sharing is Caring: GPU Sharing and CDI in Device Plugins  – https://sched.co/1YeQ7 – https://youtu.be/Q2GuTUO170w

## Issues and Contributing

[Check out the Contributing document!](CONTRIBUTING.md)

* Please let us know by [filing a new issue](https://github.com/cncf-tags/container-device-interface/issues/new)
* You can contribute by opening a [pull request](https://help.github.com/articles/using-pull-requests/)
