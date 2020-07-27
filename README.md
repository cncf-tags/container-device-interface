# CDI - The Container Device Interface
## What is CDI?

CDI (Container Device Interface), is a specification, for container runtimes, to support third party devices.

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

## Examples
```bash
$ mkdir /etc/cdi
$ cat > /etc/cdi/vendor.json <<EOF
{
  "cdiVersion": "0.2.0",
  "kind": "vendor.com/device",
  "devices": [
    {
      "name": "myDevice",
      "ociEdits": {
        "deviceNodes": [
          {"hostPath": "/dev/card1", "containerPath": "/dev/card1"}
          {"hostPath": "/dev/card-render1", "containerPath": "/dev/card-render1"}
        ]
      }
    }
  ],
  "ociEdits": {
    "deviceNodes": [
      {"hostPath": "/dev/vendorctl", "containerPath": "/dev/vendorctl"}
    ],
    "mounts": [
      {"hostPath": "/bin/vendorBin", "containerPath": "/bin/vendorBin"},
      {"hostPath": "/usr/lib/libVendor.so.0", "containerPath": "/usr/lib/libVendor.so.0"}
    ],
    "hooks": [
      {"create-container": {"path": "/bin/vendor-hook"} },
      {"start-container": {"path": "/usr/bin/ldconfig"} }
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
