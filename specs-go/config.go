package specs

// Spec is the base configuration for CDI
type Spec struct {
	Version          string   `json:"cdiVersion"`
	Kind             string   `json:"kind"`
	KindShort        []string `json:"kindShort,omitempty"`
	ContainerRuntime []string `json:"containerRuntime,omitempty"`

	Devices        []Devices      `json:"devices"`
	ContainerEdits ContainerEdits `json:"containerEdits,omitempty"`
}

// Devices is a "Device" a container runtime can add to a container
type Devices struct {
	Name           string         `json:"name"`
	NameShort      []string       `json:"nameShort"`
	ContainerEdits ContainerEdits `json:"containerEdits"`
}

// ContainerEdits are edits a container runtime must make to the OCI spec to expose the device.
type ContainerEdits struct {
	Env         []string      `json:"env,omitempty"`
	User        *User         `json:"user,omitempty"`
	DeviceNodes []*DeviceNode `json:"deviceNodes,omitempty"`
	Hooks       []*Hook       `json:"hooks,omitempty"`
	Mounts      []*Mount      `json:"mounts,omitempty"`
}

// DeviceNode represents a device node that needs to be added to the OCI spec.
type DeviceNode struct {
	HostPath      string   `json:"hostPath"`
	ContainerPath string   `json:"containerPath"`
	Permissions   []string `json:"permissions,omitempty"`
}

// User represents a user the container process runs as
type User struct {
	UID            uint32   `json:"uid"`
	GID            uint32   `json:"gid"`
	Umask          uint32   `json:"umask,omitempty"`
	AdditionalGids []uint32 `json:"additionalgids,omitempty"`
}

// Mount represents a mount that needs to be added to the OCI spec.
type Mount struct {
	HostPath      string   `json:"hostPath"`
	ContainerPath string   `json:"containerPath"`
	Options       []string `json:"options,omitempty"`
}

// Hook represents a hook that needs to be added to the OCI spec.
type Hook struct {
	HookName string   `json:"hookName"`
	Path     string   `json:"path"`
	Args     []string `json:"args,omitempty"`
	Env      []string `json:"env,omitempty"`
	Timeout  *int     `json:"timeout,omitempty"`
}
