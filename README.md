# CDI - The Container Device Interface
## What is CDI?

CDI (Container Device Interface), is a [specification](SPEC.md), for container runtimes, to support third party devices.

CDI concerns itself only with enabling container to be device aware. Areas like resource management are explicitly left out of CDI (and is expected to be handled by the orchestrator). Because of this focus, the CDI specification is simple to implement and allows great flexibility to runtimes and orchestrators.

Note: The CDI model is based on the Container Networking Interface (CNI) model and specification.

## Why is CDI needed?

On Linux, enabling a container to be device aware used to be as simple as exposing a device node in that container.
However, as devices and software grows more complex, vendors want to perform more operations, such as:

- Exposing a device to a container can require exposing more than one device node, mounting files from the runtime namespace or hide procfs entries.
- Performing compatibility checks between the container and the device (e.g: Can this container run on this device).
- Performing runtime specific operations (e.g: VM vs linux containers based runtimes).
- Performing device specific operations (e.g: scrubbing the memory of a GPU or reconfiguring an FPGA).

In the absence of a standard for third party devices, vendors often have to write and maintain multiple plugins for different runtimes or even directly contribute vendor specific code in the runtime.
Additionally runtimes don't uniformly expose a plugin system (or even expose a plugin system at all) leading to duplication of the functionality in higher level abstractions (such as Kubernetes device plugins).

## How does CDI work?

For CDI to work the following needs to be done:

- CDI file containing update for the OCI spec in JSON format should be present in the CDI
  spec directory. Default directories are /etc/cdi and /var/run/cdi

- Fully qualified device name should be passed to the runtime either
  using command line parameters for podman or using container annotations
  for CRI-O and Containerd

- Container runtime should be able to find CDI file by the device name
  and update container config using CDI file content.

## How to configure CDI?

### CRI-O configuration

In CRI-O CDI support is enabled by default. It is configured with the default `/etc/cdi, /var/run/cdi`
CDI directory locations. Therefore you can start using CDI simply by dropping CDI configuration files
in either of those directories, static configuration into `/etc/cdi` and dynamically updated one into
`/var/run/cdi`. If you are unsure of the configured directories you can run this command to find them
out:

```bash
$ crio config |& grep -B1 -A5 cdi_spec_dirs
```

### Containerd configuration

To enable and configure CDI support in the [containerd runtime](https://github.com/containerd/containerd) 2 configuration options `enable_cdi` and `cdi_spec_dirs` should be set in the `plugins."io.containerd.grpc.v1.cri` section of the containerd configuration file (`/etc/containerd/config.toml` by default):

```
[plugins."io.containerd.grpc.v1.cri"]
  enable_cdi = true
  cdi_spec_dirs = ["/etc/cdi", "/var/run/cdi"]
```

Remember to restart containerd for any configuration changes to take effect.

### Podman configuration

[podman](https://github.com/containers/podman) does not require any specific configuration to enable CDI support and processes specified `--device` flags directly. If fully-qualified device selectors (e.g. `vendor.com/device=myDevice`) are included the CDI specifications at the default location (`/etc/cdi` and `/var/run/cdi`) are checked for matching devices.

*Note:* Although initial support was added in [`v3.2.0`](https://github.com/containers/podman/releases/tag/v3.2.0) this was updated for the tagged `v0.3.0` CDI spec in [`v4.1.0-rc.1`](https://github.com/containers/podman/releases/tag/v4.1.0-rc1) with [commit a234e4e](https://github.com/containers/podman/commit/a234e4e19662e172472877ce69523f4afea5c12e).

## Examples
```bash
$ mkdir /etc/cdi
$ cat > /etc/cdi/vendor.json <<EOF
{
  "cdiVersion": "0.4.0",
  "kind": "vendor.com/device",
  "devices": [
    {
      "name": "myDevice",
      "containerEdits": {
        "deviceNodes": [
          {"path": "/dev/card1", "type": "c", "major": 25, "minor": 25, "fileMode": 384, "permissions": "rw", "uid": 1000, "gid": 1000},
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

# CLI examples below

# Verbose
$ docker/podman run --device vendor.com/device=myDevice --device vendor.com/device=myDevice2 ...

# Less verbose, through infering the vendor from the device name
$ docker/podman run --device myDevice ...

# Special case
$ docker/podman run --device vendor.com/device=all ...
```

## Issues and Contributing

[Checkout the Contributing document!](CONTRIBUTING.md)

* Please let us know by [filing a new issue](https://github.com/RenaudWasTaken/cdi/issues/new)
* You can contribute by opening a [pull request](https://help.github.com/articles/using-pull-requests/)
