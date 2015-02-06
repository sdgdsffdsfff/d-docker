package dockerapi

import (
)

type Protocol string

type Port struct {
	// Optional: If specified, this must be a DNS_LABEL.  Each named port
	// in a pod must have a unique name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Optional: If specified, this must be a valid port number, 0 < x < 65536.
	HostPort int `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	// Required: This must be a valid port number, 0 < x < 65536.
	ContainerPort int `json:"containerPort" yaml:"containerPort"`
	// Optional: Supports "TCP" and "UDP".  Defaults to "TCP".
	Protocol Protocol `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	// Optional: What host IP to bind the external port to.
	HostIP string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
}

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	// Required: This must be a C_IDENTIFIER.
	Name string `json:"name" yaml:"name"`
	// Optional: defaults to "".
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	// Required: This must match the Name of a Volume [above].
	Name string `json:"name" yaml:"name"`
	// Optional: Defaults to false (read-write).
	ReadOnly bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	DstPath  string `json:"dstPath,omitempty" yaml:"dstPath,omitempty"`
	// Required.
	MountPath string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
}

// IntOrString is a type that can hold an int or a string.  When used in
// JSON or YAML marshalling and unmarshalling, it produces or consumes the
// inner type.  This allows you to have, for example, a JSON field that can
// accept a name or number.
type IntOrString struct {
	Kind   IntstrKind
	IntVal int
	StrVal string
}

// IntstrKind represents the stored type of IntOrString.
type IntstrKind int

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// Optional: Path to access on the HTTP server.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Required: Name or number of the port to access on the container.
	Port IntOrString  `json:"port,omitempty" yaml:"port,omitempty"`
	// Optional: Host name to connect to, defaults to the pod IP.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket
type TCPSocketAction struct {
	// Required: Port to connect to.
	Port IntOrString `json:"port,omitempty" yaml:"port,omitempty"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem.  The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work.  To use
	// a shell, you need to explicitly call out to that shell.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
}
// LivenessProbe describes a liveness probe to be examined to the container.
// TODO: pass structured data to the actions, and document that data here.
type LivenessProbe struct {
	// HTTPGetProbe parameters, required if Type == 'http'
	HTTPGet *HTTPGetAction `yaml:"httpGet,omitempty" json:"httpGet,omitempty"`
	// TCPSocketProbe parameter, required if Type == 'tcp'
	TCPSocket *TCPSocketAction `yaml:"tcpSocket,omitempty" json:"tcpSocket,omitempty"`
	// ExecProbe parameter, required if Type == 'exec'
	Exec *ExecAction `yaml:"exec,omitempty" json:"exec,omitempty"`
	// Length of time before health checking is activated.  In seconds.
	InitialDelaySeconds int64 `yaml:"initialDelaySeconds,omitempty" json:"initialDelaySeconds,omitempty"`
}

// Handler defines a specific action that should be taken
// TODO: pass structured data to these actions, and document that data here.
type Handler struct {
	// One and only one of the following should be specified.
	// Exec specifies the action to take.
	Exec *ExecAction `json:"exec,omitempty" yaml:"exec,omitempty"`
	// HTTPGet specifies the http request to perform.
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
}

// Lifecycle describes actions that the management system should take in response to container lifecycle
// events.  For the PostStart and PreStop lifecycle handlers, management of the container blocks
// until the action is complete, unless the container process fails, in which case the handler is aborted.
type Lifecycle struct {
	// PostStart is called immediately after a container is created.  If the handler fails, the container
	// is terminated and restarted.
	PostStart *Handler `json:"postStart,omitempty" yaml:"postStart,omitempty"`
	// PreStop is called immediately before a container is terminated.  The reason for termination is
	// passed to the handler.  Regardless of the outcome of the handler, the container is eventually terminated.
	PreStop *Handler `yaml:"preStop,omitempty" json:"preStop,omitempty"`
}

// PullPolicy describes a policy for if/when to pull a container image
type PullPolicy string

// Container represents a single container that is expected to be run on the host.
type Container struct {
	ID	string	`json:"ID" yaml:"ID"`
	// Required: This must be a DNS_LABEL.  Each container in a pod must
	// have a unique name.
	Name string `json:"name" yaml:"name"`
	// Required.
	Image string `json:"image" yaml:"image"`
	// Optional: Defaults to whatever is defined in the image.
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
	// Optional: Defaults to Docker's default.
	WorkingDir string   `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	Ports      []Port   `json:"ports,omitempty" yaml:"ports,omitempty"`
	Env        []EnvVar `json:"env,omitempty" yaml:"env,omitempty"`
	// Optional: Defaults to unlimited.
	Memory int `json:"memory,omitempty" yaml:"memory,omitempty"`
	// Optional: Defaults to unlimited.
	CPU           int            `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	VolumeMounts  []VolumeMount  `json:"volumeMounts,omitempty" yaml:"volumeMounts,omitempty"`
	LivenessProbe *LivenessProbe `json:"livenessProbe,omitempty" yaml:"livenessProbe,omitempty"`
	Lifecycle     *Lifecycle     `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	// Optional: Defaults to /dev/termination-log
	TerminationMessagePath string `json:"terminationMessagePath,omitempty" yaml:"terminationMessagePath,omitempty"`
	// Optional: Default to false.
	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	// Optional: Policy for pulling images for this container
	ImagePullPolicy PullPolicy `json:"imagePullPolicy" yaml:"imagePullPolicy"`
}

