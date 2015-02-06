package starting

import (
	"github.com/cloudfoundry/yagnats"
	steno "github.com/cloudfoundry/gosteno"
	"encoding/json"
	"dea-docker/src/dea/config"
)

const (
	NATS_SUBJECT_DROPLET_EXIT = "droplet.exited"
	NATS_SUBJECT_HEARTBEAT    = "dea.heartbeat"
	NATS_SUBJECT_JAE_STARTED  = "jae.started"
	NATS_SUBJECT_JAE_STOPPED  = "jae.stopped"
	NATS_SUBJECT_ROUTER_REGISTER = "router.register"
	NATS_SUBJECT_ROUTER_UNREGISTER = "router.unregister"
	
)

func SendRouterUnRegister(messageBus yagnats.NATSClient,logger *steno.Logger, ins *Instance, conf *config.Config ){
	msg := make(map[string]interface{})
	msg["dea"] = conf.Dea.Uuid
	msg["app"] = ins.ApplicationId
	msg["uris"] = ins.ApplicationUris
	msg["host"] = conf.Dea.LocalIp
	msg["port"] = ins.ContainerInfo.HostPort
	msg["tags"] = "dea-"+conf.Dea.Index
	msg["private_instance_id"] = ins.InstanceId
	
	Publish(messageBus, logger, NATS_SUBJECT_ROUTER_UNREGISTER, msg)
}

func SendRouterRegister(messageBus yagnats.NATSClient,logger *steno.Logger, ins *Instance, conf *config.Config ) {
	msg := make(map[string]interface{})
	msg["dea"] = conf.Dea.Uuid
	msg["app"] = ins.ApplicationId
	msg["uris"] = ins.ApplicationUris
	msg["host"] = conf.Dea.LocalIp
	msg["port"] = ins.ContainerInfo.HostPort
	msg["tags"] = "dea-"+conf.Dea.Index
	msg["private_instance_id"] = ins.InstanceId
	
	Publish(messageBus, logger, NATS_SUBJECT_ROUTER_REGISTER, msg)
}


func SendJaeStopped(messageBus yagnats.NATSClient,logger *steno.Logger, ins *Instance, conf *config.Config) {
	msg := make(map[string]interface{})
	msg["app_guid"] = ins.ApplicationId
	msg["app_name"] = ins.ApplicationName
	msg["ip"] = conf.Dea.LocalIp
	
	Publish(messageBus, logger, NATS_SUBJECT_JAE_STOPPED, msg)
}

func SendJaeStarted(messageBus yagnats.NATSClient,logger *steno.Logger, ins *Instance, conf *config.Config) {
	msg := make(map[string]interface{})
	msg["app_guid"] = ins.ApplicationId
	msg["app_name"] = ins.ApplicationName
	msg["ip"] = conf.Dea.LocalIp
	msg["log_path"] = conf.Dea.Uuid
	msg["build_pack"] = ins.BuildPack
	
	Publish(messageBus, logger, NATS_SUBJECT_JAE_STARTED, msg)
}

func SendHearthbate(messageBus yagnats.NATSClient,logger *steno.Logger,registry *InstanceRegistry ,deaId string){
	instances := registry.AllInstances() 
	insData := []interface{}{}
	
	for _, ins := range instances {
		d := make(map[string]interface{})
		d["cc_partition"] = ins.CcPartition
		d["droplet"] = ins.ApplicationId
		d["version"] = ins.Version
		d["instance"] = ins.InstanceId
		d["index"] = ins.Index
		d["state"] = ins.State
		d["state_timestamp"] = ins.Attr["state_time"]
		
		insData = append(insData,d)
	}
	
	message := make(map[string]interface{})
	
	message["droplets"] = insData
	message["dea"]	= deaId
	
	Publish(messageBus, logger, NATS_SUBJECT_HEARTBEAT, message)
	
	
}

func SendCrashedMessage(messageBus yagnats.NATSClient,logger *steno.Logger,ins *Instance, reason string) {
	
	var msg = make(map[string]interface{})
	msg["cc_partition"] = ins.CcPartition
	msg["droplet"] = ins.ApplicationId
	msg["version"] = ins.Version
	msg["instance"] = ins.InstanceId
	msg["index"] = ins.Index
	msg["reason"] = reason
	msg["exit_status"] = ins.ExitStatus
	msg["exit_description"] = ins.ExitDescription
	msg["crash_timestamp"] = ins.ExitTime
	
	Publish(messageBus, logger, NATS_SUBJECT_DROPLET_EXIT, msg)
}

func Publish(messageBus yagnats.NATSClient,logger *steno.Logger, subject string, data interface{}) {
	
	//解析数据
	d ,err := json.Marshal(data)
	
	if err != nil {
		logger.Errorf("publish:%s,data:%s	fail,%s",subject,string(d), err)
	}
	
	err = messageBus.Publish(subject, d)
	
	if err != nil {
		logger.Errorf("publish:%s,data:%s	fail,%s",subject,string(d), err)
	}
	logger.Infof("publish:%s,data:%s	success",subject,string(d))
}