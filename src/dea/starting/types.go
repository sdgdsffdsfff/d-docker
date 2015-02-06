package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/dockerapi"
	"dea-docker/src/dea/config"
	"sync"
)

const (
	//容器状态的一系列常量
	   STATE_BORN = "BORN"
      STATE_STARTING = "STARTING"
      STATE_RUNNING = "RUNNING"
      STATE_STOPPING = "STOPPING"
      STATE_STOPPED = "STOPPED"
      STATE_CRASHED = "CRASHED"
      STATE_DELETED = "DELETED"
      STATE_RESUMING = "RESUMING"
      STATE_EVACUATING = "EVACUATING"
)

//stop instance message
type NatsStopMsg struct {
	Droplet				string					`json:"droplet"`
	Version				string					`json:"version"`
	Instances			[]string				`json:"instances"`
	Indices				[]int					`json:"indices"`
	Isdelete			bool					`json:"isdelete"`
}

type FilterInstanceMessage struct {
	Droplet				string					
	Version				string					
	Instances			[]string				
	Indices				[]int					
	States				[]string				
}


//start msg
type NatsStartMsg struct {
	Droplet 			string					`json:"droplet"`
	Registry			string					`json:"registry"`
	Tags				map[string]string		`json:"tags"`
	Name				string					`json:"name"`
	Uris				[]string				`json:"uris"`
	Prod				bool					`json:"prod"`
	Sha1				string	   				`json:"sha1"`
	ExecutableFile		string					`json:"executableFile"`
	ExecutableUri		string					`json:"executableUri"`
	Version				string					`json:"version"`
	Services			[]string				`json:"services"`
	Limits				map[string]int			`json:"limits"`
	CcPartition			string					`json:"cc_partition"`
	Env 				[]string				`json:"env"`
	Console				bool					`json:"console"`
	Debug				string					`json:"debug"`
	BuildPack			string					`json:"build_pack"`
	Index				int						`json:"index"`
	ImageName			string					`json:"image_name"`
	ImageTarg   		string					`json:"image_targ"`
}

//应用状态事件
type Transition struct {
	From		string
	To			string
	Callback	func(instance *Instance)
}

//instance obj
type Instance struct {
	docker				*dockerapi.Docker
	logger				*steno.Logger
	portPool			*PortPool
	conf				*config.Config
	events				[]*Transition
	healthCheck			*HealthCheck
	lock 				sync.Mutex
	ApplicationId		string //app guid
	ApplicationName		string
	BuildPack			string
	InstanceId			string //app instance id in vm
	State				string	//实例状态
	StartMessage		*NatsStartMsg						`json:"start_message"`
	Attr				map[string]string//扩展属性
	CcPartition			string
	Version				string
	Index				int
	ApplicationUris		[]string
	ExitStatus			string
	ExitDescription		string
	ExitTime			string
	ContainerInfo		*ContainerInfo						`json:"container_info"`
}


//容器基本信息
type ContainerInfo struct {

	Id					string 						`json:"container_id"`
	AufsPath			string 						`json:"container_aufs_path"`
	ContainerPath		string						`json:"container_path"`
	StateTime			string						`json:"container_time"`
	Running				bool						`json:"container_running"`
	HostPort			int							`json:"host_port"`
	HostIp				string						`json:"host_ip"`
}

