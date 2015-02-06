package api

import (
	steno "github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/yagnats"
	"encoding/json"
	"fmt"
	"os"
)

const (
	SUBJECT_ROUTER_START = "router.start"
	SUBJECT_DEA_UID_START  = "dea.%s.start"
	SUBJECT_DEA_UID_STOP	= "dea.%s.stop"
	SUBJECT_DEA_STOP	= "dea.stop"
	SUBJECT_DEA_UPDATE = "dea.update"
	SUBJECT_DEA_FIND_DROPLET = "dea.find.droplet"
	
)

type Nats struct {
	messageBus        yagnats.NATSClient
	handle 			*NatsMessageHandle
	logger				*steno.Logger
	uuid				string
}

func NewNats(mbus yagnats.NATSClient, msgHandle *NatsMessageHandle, uid string) *Nats {
	return &Nats{
		messageBus:	mbus,
		handle:	msgHandle,
		logger:	steno.NewLogger("dea-docker"),
		uuid:	uid,
	}
}

//启动
func (n *Nats) Start() {

	var err error
	
	//router启动的探测事件
	_, err = n.subscribe(SUBJECT_ROUTER_START, func(message *yagnats.Message, subject string){
		n.handle.HandleRouterStart(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	//实例启动事件
	_, err = n.subscribe(fmt.Sprintf(SUBJECT_DEA_UID_START,n.uuid), func(message *yagnats.Message, subject string){
		n.handle.HandleDeaStart(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	//实例停止 事件
	_, err = n.subscribe(fmt.Sprintf(SUBJECT_DEA_UID_STOP,n.uuid), func(message *yagnats.Message, subject string){
		n.handle.HandleDeaStopApp(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	//实例停止 事件
	_, err = n.subscribe(SUBJECT_DEA_STOP, func(message *yagnats.Message, subject string){
		n.handle.HandleDeaStop(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	//实例uri更新 
	_, err = n.subscribe(SUBJECT_DEA_UPDATE, func(message *yagnats.Message, subject string){
		n.handle.HandleDeaUpdateApp(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	//实例查询
	_, err = n.subscribe(SUBJECT_DEA_FIND_DROPLET, func(message *yagnats.Message, subject string){
		n.handle.HandleDeaFindApp(message)
	})
	if err != nil {
		n.logger.Errorf("start nats fail,%s", err)
		os.Exit(1)
	}
	
	n.logger.Infof("Nats start and subscribe success")
}


//register msg
func (n *Nats) subscribe(subject string, successCall func(*yagnats.Message, string)) (int, error) {
	callback := func(message *yagnats.Message) {
		n.logger.Infof("subscribe:%s,data:%s",subject,message.Payload)
		successCall(message, message.ReplyTo)
	}
	
	sid, err := n.messageBus.Subscribe(subject, callback)
	
	if err != nil {
		return -1, err
	}
	
	n.logger.Infof("nats subscribe:%s",subject)
	return sid,nil
}

//push msg
func (n *Nats) publish(subject string, data interface{}) error {
	
	//解析数据
	d ,err := json.Marshal(data)
	
	if err != nil {
		return err
	}
	n.logger.Infof("publish:%s,data:%s",subject,string(d))
	
	err = n.messageBus.Publish(subject, d)
	
	if err != nil {
		return err
	}
	
	return nil
}