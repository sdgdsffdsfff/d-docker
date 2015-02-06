package starting

import (
	"sync"
	steno "github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/yagnats"
	"dea-docker/src/dea/config"
	"time"
)

const perfix = "InstanceRegistry "

//instance registry
type InstanceRegistry struct {

	instances 			map[string] *Instance //实例基本信息
	instanceById 		map[string]map[string]*Instance//以应用ID 为基准存放实例信息
	lock 				sync.Mutex
	tickerLock			sync.Mutex
	ticker				*time.Ticker
	logger				*steno.Logger
	messageBus	        yagnats.NATSClient
	conf				*config.Config
	
}

func NewInstanceRegistry (mbus yagnats.NATSClient, conf *config.Config) *InstanceRegistry {
	return &InstanceRegistry{
		instances:			make(map[string]*Instance),
		instanceById:		make(map[string]map[string]*Instance),
		logger:				steno.NewLogger("dea-docker"),
		messageBus:			mbus,
		conf:				conf,
	}
}


func (i *InstanceRegistry) Start () {
	
	i.tickerLock.Lock()
	i.ticker = time.NewTicker(time.Second * 5)
	
	i.tickerLock.Unlock()
	
	go func() {
			for {
				select {
				case <-i.ticker.C:
					i.RegistryRouter()
				}
			}
		}()
}


func (i *InstanceRegistry) FilterInstanceByApplication(applicationId string) []interface{} {
	
	instances := i.AllInstances() 
	
	insData := []interface{}{}
	
	for _, ins := range instances {
		
		if ins.ApplicationId != applicationId {
			continue	
		}
		d := make(map[string]interface{})
		d["start_message"] = ins.StartMessage
		d["container_info"] = ins.ContainerInfo
		insData = append(insData,d)
	}
	
	return insData
}


//registry to router
func (i *InstanceRegistry) RegistryRouter() {
	instances := i.AllInstances()
	
	for _,instance := range instances {
		if instance.ContainerInfo == nil {
			continue	
		}
		
		if instance.State != STATE_RUNNING {
			continue	
		}
		
		if !instance.ContainerInfo.Running {
			continue	
		}
		SendRouterRegister(i.messageBus, i.logger, instance, i.conf)
	}
}


//查询instance
func (i *InstanceRegistry) FilterInstance(message FilterInstanceMessage, call func(instance *Instance) ) {
	appId := message.Droplet
	
	intances, found := i.instanceById[appId]
	
	if !found {
		i.logger.Errorf("Instance_Registry Filter Instance but appId:{%s} not found", appId)
	}
	
	if intances == nil || len(intances) <=0 {
		i.logger.Errorf("Instance_Registry Filter Instance but appId:{%s} not found", appId)
	}
	
	version 	:= message.Version
	indices 	:= message.Indices
	states  	:= message.States
	instanceIds := message.Instances
	
	
	for _, instance := range intances {
		
		if instance.Version != version {
			continue	
		}
		
		instanceIdFlag := true
		for _,instanceId := range instanceIds {
			if instance.InstanceId == instanceId {
				instanceIdFlag = true
				break	
			}
			instanceIdFlag = false
		}
		
		if !instanceIdFlag {
			continue
		}
		
		indiceFlag := true
		for _, indice := range indices {
			if instance.Index == indice {
				indiceFlag = true
				break
			} 
			indiceFlag = false
		}
		
		if !indiceFlag {
			continue	
		} 
		
		stateFlag := true
		for _,state := range states {
			if instance.State == state {
				stateFlag = true	
			}
			stateFlag = false
		}
		
		if !stateFlag {
			continue	
		} 
		
		i.logger.Infof("Execute FilterInstance Call,(appId:%s,InstanceId:%s,index:%s,version:%s)", appId, instance.InstanceId, instance.Index, instance.Version)
		call(instance)
	}
}

//注册instance
func (i *InstanceRegistry) Register(instance *Instance) {
	i.addInstance(instance)
}

//取消注册instance
func (i *InstanceRegistry) UnRegister(instance *Instance) {
	i.removeInstance(instance)
}

//return all instances
func (i *InstanceRegistry) AllInstances() map[string]*Instance {
	i.lock.Lock()
	defer i.lock.Unlock()
	
	instances := i.instances
	
	return instances
}

func (i *InstanceRegistry) addInstance(instance *Instance) {
	
	i.lock.Lock()
	defer i.lock.Unlock()
	
	app_id := instance.ApplicationId
	instanceId := instance.InstanceId
	
	i.logger.Infof("%s addInstance ,appId:%s, instanceId:%s", perfix, app_id, instanceId)
	
	i.instances[instanceId]	= instance
	
	ins, found := i.instanceById[app_id]
	if !found {
		ins = make(map[string]*Instance)
	}
	
	ins[instanceId] = instance
	
	i.instanceById[app_id] = ins
	i.logger.Infof("%s addInstance ,appId:%s, instanceId:%s	 success", perfix, app_id, instanceId)
}

func (i *InstanceRegistry) removeInstance (instance *Instance) {
	
	i.lock.Lock()
	defer i.lock.Unlock()
	
	app_id := instance.ApplicationId
	instanceId := instance.InstanceId
	
	i.logger.Infof("%s removeInstance ,appId:%s, instanceId:%s", perfix, app_id, instanceId)
	
	_, found := i.instances[instanceId]
	
	if found {
		delete(i.instances, instanceId)
	}
	
	ins, found := i.instanceById[app_id]
	
	if found {
		_, found = ins[instanceId]
		
		if found {
			delete(ins, instanceId)
		}
		
		if len(ins) == 0 {
			delete(i.instanceById, app_id)
		}else {
			i.instanceById[app_id] = ins
		}
	}
	
	i.logger.Infof("%s removeInstance ,appId:%s, instanceId:%s	 success", perfix, app_id, instanceId)
}