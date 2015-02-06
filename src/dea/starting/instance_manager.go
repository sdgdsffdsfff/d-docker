package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/yagnats"
	"errors"
	"encoding/json"
	"dea-docker/src/dea/util"
	"dea-docker/src/dea/config"
	"dea-docker/src/dea/dockerapi"
)

//instance manager
type InstanceManager struct {
	logger					*steno.Logger
	messageBus	      	 	yagnats.NATSClient
	resourceManager			*ResourceManager
	instanceRegistry 		*InstanceRegistry
	snapshot				*Snapshot
	deaId					string
	conf					*config.Config
	docker					*dockerapi.Docker
	portPool				*PortPool
}

//create instance manager
func NewInstanceManager(mbus yagnats.NATSClient, resManager *ResourceManager, instanceRegistry *InstanceRegistry ,snapshot	*Snapshot, conf *config.Config) (*InstanceManager,error) {
	
	docker, err := dockerapi.NewDocker(conf)
	
	if err != nil {
		return nil,err	
	}
	
	portPool, err := NewPortPool(conf)
	
	if err != nil {
		return nil, err	
	}
	
	return &InstanceManager{
		logger:				steno.NewLogger("dea-docker"),
		messageBus:			mbus,
		resourceManager:	resManager,
		instanceRegistry:	instanceRegistry,
		snapshot:			snapshot,
		deaId:				conf.Dea.Uuid,
		conf:				conf,
		docker:				docker,
		portPool:			portPool,
	},nil
}

func (i *InstanceManager) CreateInstance(startmsg *NatsStartMsg) (*Instance, error) {
	
	i.printStartMsg(startmsg)
	
	if startmsg == nil {
		return nil ,errors.New("Instance Manager CreateInstance err, startMessage is empty")
	}
	
	//use default docker registry
	if startmsg.Registry == "" {
		startmsg.Registry = i.conf.Docker.Registry
	}
	
	instance , err := NewInstance(startmsg, i.docker, i.portPool, i.conf)
	if err != nil {
		return nil, err
	}
	
	instance.On(&Transition{
		From: STATE_BORN,
		To :  STATE_CRASHED,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_BORN, STATE_CRASHED)
			SendCrashedMessage(i.messageBus, i.logger,ins,"CRASHED")
			i.instanceRegistry.UnRegister(ins)
		},
	})
	//验证参数
	err = instance.Validate()
	if err != nil {
		return nil, err
	}
	
	//验证vm 资源是否足够
	memory_limit	:= int(instance.StartMessage.Limits["mem"])
	disk_limit		:= int(instance.StartMessage.Limits["disk"])
	err = i.resourceManager.ValidateResource(memory_limit, disk_limit)
	
	if err != nil {
		i.logger.Errorf("create instance:%s,check resource (memory_limit:%s,disk_limit:%s) fail,%s",instance.ApplicationId, memory_limit, disk_limit, err)
		i.crashInstance(instance, err,STATE_CRASHED)
		return nil, err
	}
	
	//初始化基础数据
	err = instance.Setup()
	if err != nil {
		i.logger.Errorf("create instance:%s,instance setup fail,%s",instance.ApplicationId, err)
		i.crashInstance(instance, err,STATE_CRASHED)
		return nil, err
	}
	
	//注册事件
	instance.On(&Transition{
		From: STATE_STARTING,
		To :  STATE_CRASHED,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_STARTING, STATE_CRASHED)
			SendCrashedMessage(i.messageBus, i.logger,ins,"CRASHED")
			i.instanceRegistry.UnRegister(ins)
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_RESUMING,
		To :  STATE_CRASHED,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_RESUMING, STATE_CRASHED)
			SendCrashedMessage(i.messageBus, i.logger,ins,"CRASHED")
			i.instanceRegistry.UnRegister(ins)
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_STARTING,
		To :  STATE_RUNNING,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_STARTING, STATE_RUNNING)
			//发送heartbat 数据
			SendHearthbate(i.messageBus, i.logger, i.instanceRegistry, i.deaId)
			// 发送jae.started 数据
			SendJaeStarted(i.messageBus, i.logger, instance, i.conf)
			//发送 router.register数据
			SendRouterRegister(i.messageBus, i.logger, ins, i.conf)
			//注册registry
			i.instanceRegistry.Register(instance)
			//save snapshot 
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_RESUMING,
		To :  STATE_RUNNING,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_RESUMING, STATE_RUNNING)
			//发送heartbat 数据
			SendHearthbate(i.messageBus, i.logger, i.instanceRegistry, i.deaId)
			// 发送jae.started 数据
			SendJaeStarted(i.messageBus, i.logger, instance, i.conf)
			//发送 router.register数据
			SendRouterRegister(i.messageBus, i.logger, ins, i.conf)
			//注册registry
			i.instanceRegistry.Register(instance)
			//save snapshot 
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_RUNNING,
		To :  STATE_CRASHED,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_RUNNING, STATE_CRASHED)
			//取消注册 registry
			i.instanceRegistry.UnRegister(instance)
			//发送 crashed message
			SendCrashedMessage(i.messageBus, i.logger, ins,"CRASHED")
			//send router.unregister
			SendRouterUnRegister(i.messageBus, i.logger, ins, i.conf)
			//save snapshot 
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_RUNNING,
		To :  STATE_STOPPING,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_RUNNING, STATE_STOPPING)
			//取消注册 registry
			i.instanceRegistry.UnRegister(instance)
			//save snapshot 
			i.snapshot.Save()
		},
	})
	
	instance.On(&Transition{
		From: STATE_STOPPING,
		To :  STATE_STOPPED,
		Callback: func (ins *Instance) {
			i.logger.Infof("execute application:%s, transition:(From:%s,To:%s) callback",ins.ApplicationId, STATE_RUNNING, STATE_STOPPED)
			//取消注册 registry
			i.instanceRegistry.UnRegister(instance)
			//save snapshot 
			i.snapshot.Save()
			//发送 jae.stopped message
			SendJaeStopped(i.messageBus, i.logger, ins, i.conf)
			
			//send router.unregister
			SendRouterUnRegister(i.messageBus, i.logger, ins, i.conf)
			
			//destroy instance
			ins.Destroy()
		},
	})
	
	//registry
	i.instanceRegistry.Register(instance)
	
	return instance,nil
}

func(i *InstanceManager) crashInstance(instance *Instance, err error,state string) {
		instance.ExitStatus = "400"
		instance.ExitDescription = err.Error()
		instance.ExitTime = util.NowTime()
		instance.SetState(state)
}

func (i *InstanceManager) printStartMsg(startmsg *NatsStartMsg) {
	
	b,err := json.Marshal(startmsg)
	
	if err != nil {
		i.logger.Errorf("printStartmsg fail, %s", err)
	}
	
	i.logger.Infof("CreateInstance message:%s", string(b))
}
