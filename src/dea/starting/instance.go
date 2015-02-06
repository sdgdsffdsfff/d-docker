package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/util"
	"dea-docker/src/dea/dockerapi"
	"dea-docker/src/dea/config"
	"errors"
	"fmt"
	"path"
	"strconv"
)

const(
		aufs_base_path = "aufs/diff"
		container_base_path = "containers"
	)

//创建实例
func NewInstance(startMessage *NatsStartMsg, docker *dockerapi.Docker, portPool *PortPool, cfg *config.Config) (*Instance , error) {
	
	if startMessage == nil {
		return nil,errors.New("NewInstance error startmessage is empty")
	}
	guid ,err := util.GetGuid()
	
	if err != nil {
		return nil, err
	}
	
	instance := &Instance{
		logger:				steno.NewLogger("dea-docker"),
		conf:				cfg,
		events:				[]*Transition{},
		InstanceId:			guid,
		ApplicationName:	startMessage.Name,
		BuildPack:			startMessage.BuildPack,
		ApplicationUris:    startMessage.Uris,
		State:				STATE_BORN,
		Attr:				make(map[string]string ),
		ApplicationId:		startMessage.Droplet,
		StartMessage:		startMessage,
		docker:				docker,
		portPool:			portPool,
	}
	
	return instance,nil
}


//创建一个事件对象
func NewTransition (from string, to string, call func(instance *Instance) )(*Transition, error) {
	
	return &Transition{
		From:	from,
		To:		to,
		Callback:	call,
	},nil
}

//启动
func (i *Instance) Start() {
	i.loggMsg("Instance begin starting","info")
	
	i.SetState(STATE_STARTING)
	
	//check image
	err := i.checkOrPullImage()
	if err != nil {
		i.loggMsg(fmt.Sprintf("Instance start fail,%s", err.Error() ), "error")
		i.SetState(STATE_CRASHED)		
		return
	}
	
	//create container and run container
	opts, err := NewCreateContainerOpts(i.StartMessage, i.portPool, i.InstanceId)
	if err != nil {
		i.loggMsg(fmt.Sprintf("Instance start fail,%s", err.Error() ), "error")
		//remove image
		i.docker.RemoveImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
		i.SetState(STATE_CRASHED)
		return
	}
	
	container , err := i.docker.RunContainer(opts, "")
	if err != nil {
		i.loggMsg(fmt.Sprintf("Instance start fail,%s", err.Error() ) , "error")
		//remove image
		i.docker.RemoveImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
		i.SetState(STATE_CRASHED)	
		return
	}
	
	if container == nil {
		i.loggMsg("Instance start fail,return container is nil", "error")
		//remove image
		i.docker.RemoveImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
		i.SetState(STATE_CRASHED)	
		return
	}
	
	//setup 
	err = i.startingCompletion(container.ID)
	if err != nil {
		i.ContainerInfo = &ContainerInfo{}
		i.loggMsg(fmt.Sprintf("Instance start fail,%s", err), "error")
		//remove image
		i.docker.RemoveImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
		//remove container
		i.docker.RemoveContainer(container.ID)
		
		i.SetState(STATE_CRASHED)	
		return
	}
	
}

//停止
func (i *Instance) Stop() error {
	
	i.loggMsg("Instance Begin Stoping", "info")
	i.SetState(STATE_STOPPING)
	
	//stop container
	err := i.docker.StopContainer(i.ContainerInfo.Id)
	
	if err != nil {
		return err	
	}
	
	i.SetState(STATE_STOPPED)
	i.loggMsg("Instance Stopped", "info")
	
	return nil
}

func (i *Instance) Destroy() {
	//remove container
	err := i.docker.RemoveContainer(i.ContainerInfo.Id)
	
	if err != nil {
		i.loggMsg("Destory Container fail:", err.Error() )
	}
	//remove images
	err = i.docker.RemoveImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
	
	if err != nil {
		i.loggMsg("Instance Stoping And Remove Image fail:%s", "error", err.Error() )
	}
}


func (i *Instance) startingCompletion (containerId string) error {
	
	container, err := i.docker.ContainerInfo(containerId)
	if err != nil {
		return err
	}
	if container == nil {
		return errors.New("InspectContainer,but container is null")
	}
	
	aufsPath := path.Join(i.conf.Docker.DockerPath, aufs_base_path, containerId)
	containerPath := path.Join(i.conf.Docker.DockerPath, container_base_path, containerId)
	var hostPort int
	hostConfig := container.HostConfig
	if hostConfig == nil {
		return errors.New("InspectContainer,but hostConfig is nil")	
	}
	
	portBindings := hostConfig.PortBindings
	if portBindings == nil {
		return errors.New("InspectContainer,but hostConfig.portBindings is nil")	
	}
	
	for _, ports := range portBindings {
		for _, port := range ports {
			hostPort,err = strconv.Atoi(port.HostPort)
			if err != nil && hostPort >0{
				break
			}
		}
	}
	
	if hostPort <=0 {
		return errors.New("can not get container host port")	
	}
	
	containerState := container.State
	
	running := containerState.Running
	if running == false {
		return errors.New("Instance Starting fail")
	}
	
	containerConfit := &ContainerInfo{
		Id:				container.ID,
		AufsPath:		aufsPath,
		ContainerPath:	containerPath,
		StateTime:		util.FormateTime(container.Created),
		Running:		running,
		HostPort:		hostPort,
		HostIp:			i.conf.Dea.LocalIp,
	}
	
	i.ContainerInfo = containerConfit
	
	//start health check
	healthCheck := NewHealthCheck()
	healthCheck.health(i)
	i.healthCheck = healthCheck
	
	i.loggMsg("Instance Starting success", "info")
	i.SetState(STATE_RUNNING)
	
	return nil
}

//检测image
//如果 image 不存在将下载image
func (i *Instance) checkOrPullImage () error {
	flag, err := i.docker.ExistsImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg)
	
	if err != nil {
		return errors.New(fmt.Sprintf("checkOrPullImage fail,%s", "error", err.Error() ) )
	}
	
	if !flag {
		err = i.docker.PullImage(i.StartMessage.ImageName, i.StartMessage.ImageTarg, i.StartMessage.Registry)
		if err != nil {
			return errors.New(fmt.Sprintf("pullImage fail.%s", "error", err.Error() ) )
		}
	}
	
	return nil
}

//参数验证
func (i *Instance) Validate() error {
	
	errPrefix := fmt.Sprintf("Create Instance(applicationId:%s) Validate fail,", i.StartMessage.Droplet)
	
	if i.StartMessage.Droplet == "" {
		return errors.New(fmt.Sprintf(errPrefix+" applicationId is empty"))
	}
	if i.StartMessage.Name == "" {
		return errors.New(fmt.Sprintf(errPrefix+" appName is empty"))
	}
	if len(i.StartMessage.Uris) == 0 {
		return errors.New(fmt.Sprintf(errPrefix+" Uris is empty"))
	}
	if i.StartMessage.ImageName == "" {
		return errors.New(fmt.Sprintf(errPrefix+" ImageName is empty"))
	}
	
	if i.StartMessage.Version == "" {
		return errors.New(fmt.Sprintf(errPrefix+" Version is empty"))
	}
	
	if len(i.StartMessage.Limits) == 0 {
		return errors.New(fmt.Sprintf(errPrefix+" Limits is empty"))
	}
	_,found := i.StartMessage.Limits["mem"]
	
	if !found {
		return errors.New(fmt.Sprintf(errPrefix+" Limits_memory_limit empty"))
	}
	
	_,found = i.StartMessage.Limits["disk"]
	if !found {
		return errors.New(fmt.Sprintf(errPrefix+" Limits_disk_limit empty"))
	}
	if i.StartMessage.CcPartition == "" {
		return errors.New(fmt.Sprintf(errPrefix+" cc_partition is empty"))
	}
	if i.StartMessage.BuildPack == "" {
		return errors.New(fmt.Sprintf(errPrefix+" BuildPack is empty"))
	}
	return nil
}

//初始化一些基本数据
func (i *Instance) Setup() error {
	
	if i.StartMessage.ImageTarg == "" || len(i.StartMessage.ImageTarg) ==0 {
		i.StartMessage.ImageTarg = "latest"
	}
	
	i.Version = i.StartMessage.Version
	i.CcPartition = i.StartMessage.CcPartition
	i.Index = i.StartMessage.Index
	
	return nil
}

//修改实例状态
func (i *Instance) SetState (state string) {
	
	if state == STATE_CRASHED || state == STATE_STOPPED || state == STATE_STOPPING {
		if i.healthCheck != nil {
			i.healthCheck.stop()
			i.healthCheck = nil	
		}
	}
	
	oldState := i.State
	ets := []*Transition{}
	//获取event
	for _, event := range i.events {
		if event.From == oldState && event.To == state {
			ets = append(ets, event)
		}
	}
	
	//修改状态 
	i.State = state
	i.Attr["state_time"] = util.NowTime()
	
	if len(ets) != 0 {
		for _, e := range ets {
			go i.emit(e)
		}
	}
	
	i.loggMsg(fmt.Sprintf("Instance update State From: %s, To: %s	success", oldState, state) ,"info")
}

//注册事件
func (i *Instance)On(event *Transition) {
	if event == nil {
		i.loggMsg("Instance On Event fail, event is empty", "error")
		return
	}
	
	i.events = append(i.events, event)
	
	i.loggMsg(fmt.Sprintf("Instance On Event[From:%s,To:%s] success", event.From, event.To) , "info")
}


//触发事件
func (i *Instance) emit(event *Transition) {
	if event == nil {
		i.loggMsg("Instance Emit event fail ,event is empty","error")
		return
	}
	
	call := event.Callback
	call(i)
}

func (i *Instance) loggMsg(msg string ,t string, a ... interface{}) {
	switch t {
		case "warn":
			i.logger.Warnf("applicationId:%s,instanceId:%s,instance_state:%s  "+msg,i.ApplicationId,i.InstanceId,i.State,a)
		case "info":
			i.logger.Infof("applicationId:%s,instanceId:%s,instance_state:%s  "+msg,i.ApplicationId,i.InstanceId,i.State,a)
		case "error":
			i.logger.Errorf("applicationId:%s,instanceId:%s,instance_state:%s  "+msg,i.ApplicationId,i.InstanceId,i.State,a)
		default:
			i.logger.Infof("applicationId:%s,instanceId:%s,instance_state:%s  "+msg,i.ApplicationId,i.InstanceId,i.State,a)
	}
}