package api 

import (
	steno "github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/yagnats"
	"dea-docker/src/dea/starting"
	"encoding/json"
)

type NatsMessageHandle struct {
	logger				*steno.Logger
	instanceManage		*starting.InstanceManager
	instancesRegistry 	*starting.InstanceRegistry
}

//创建handle实例
func NewNatsMessageHandle (insManager *starting.InstanceManager, insRegistry *starting.InstanceRegistry) *NatsMessageHandle {
	
	return &NatsMessageHandle{
		logger:				steno.NewLogger("dea-docker"),
		instanceManage: 	insManager,
		instancesRegistry:	insRegistry,
	}
}

//响应router 启动事件,重新注册vm中打所有containers
func (n *NatsMessageHandle) HandleRouterStart (message *yagnats.Message) {
	go n.instancesRegistry.RegistryRouter()
	return
}

//响应cc发送的启动消息
func (n *NatsMessageHandle) HandleDeaStart (message *yagnats.Message) {
	var startMsg starting.NatsStartMsg
	err := n.messageToObj(message, &startMsg)
	if err != nil {
		n.logger.Errorf("HandleDeaStart fail,%s", err.Error() )
		return
	}
	
	if &startMsg == nil {
		n.logger.Errorf("HandleDeaStart fail,convert startMessage nil")
		return
	}
	if n.instanceManage == nil {
		n.logger.Errorf("HandleDeaStart fail,instanceManager is nil")
		return
	}
	
	instance , err := n.instanceManage.CreateInstance(&startMsg)
	
	if err != nil {
		n.logger.Errorf("NatsMessageHandle CreateInstance fail,%s", err.Error())
		return
	}
	
	go instance.Start()
	return
}

//响应cc发送的停止消息
func (n *NatsMessageHandle) HandleDeaStopApp (message *yagnats.Message) {

	var stopMsg starting.NatsStopMsg
	
	err := n.messageToObj(message, &stopMsg)
	
	if err != nil {
		n.logger.Errorf("HandleDeaStopApp fal,%s", err.Error())	
		return
	}
	
	//stop instances
	
	filterMsg := starting.FilterInstanceMessage{
		Droplet: 			stopMsg.Droplet,					
		Version:			stopMsg.Version,					
		Instances:			stopMsg.Instances,
		Indices:			stopMsg.Indices,					
	}
	
	call := func(ins *starting.Instance) {
		err := ins.Stop()
		if err != nil {//stop fail
			n.logger.Errorf("Instance Stop fail,%s", err.Error() )
		}	
	}
	
	go n.instancesRegistry.FilterInstance(filterMsg, call)
	return
}

//响应cc发送的停止消息
func (n *NatsMessageHandle) HandleDeaStop (message *yagnats.Message) {
	var stopMsg starting.NatsStopMsg
	
	err := n.messageToObj(message, &stopMsg)
	
	if err != nil {
		n.logger.Errorf("HandleDeaStopApp fal,%s", err.Error())	
		return
	}
	
	//stop instances
	
	filterMsg := starting.FilterInstanceMessage{
		Droplet: 			stopMsg.Droplet,					
		Version:			stopMsg.Version,					
		Instances:			stopMsg.Instances,
		Indices:			stopMsg.Indices,					
	}
	
	call := func(ins *starting.Instance) {
		err := ins.Stop()
		if err != nil {//stop fail
			n.logger.Errorf("Instance Stop fail,%s", err.Error() )
		}	
	}
	
	go n.instancesRegistry.FilterInstance(filterMsg, call)
	
	return
}

//响应cc发送的update app uri消息
func (n *NatsMessageHandle) HandleDeaUpdateApp (message *yagnats.Message) {

}

//响应cc发送的查询app消息
func (n *NatsMessageHandle) HandleDeaFindApp (message *yagnats.Message) {

}

func (n *NatsMessageHandle) messageToObj(message *yagnats.Message, v interface{}) error{
		
	payload := message.Payload

	err := json.Unmarshal(payload, &v)
	
	if err != nil {
		return err
	}
	
	return nil
}
