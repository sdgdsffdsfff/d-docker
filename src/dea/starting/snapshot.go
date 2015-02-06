package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/util"
	"dea-docker/src/dea/config"
	"os"
	"io/ioutil"
	"encoding/json"
	"path"
)

const snapshotName = "instances.json"

type Instances struct {
	Instances 			[]Instance				`json:"instances"`
}

type Snapshot struct {
	logger				*steno.Logger
	instanceRegistry 	*InstanceRegistry
	instanceManager		*InstanceManager
	config				*config.Config
}

func NewSnapshot (conf *config.Config) *Snapshot{
	
	return &Snapshot{
		logger:			steno.NewLogger("dea-docker"),
		config:			conf,
	}
}

func (s *Snapshot) Configure(insRegistry *InstanceRegistry, insManager *InstanceManager) {
	s.instanceRegistry = insRegistry
	s.instanceManager = insManager
}
//save snapshot
func (s *Snapshot) Save() {
	data := make(map[string]interface{})
	nowTime := util.NowTime()
	
	instances := s.instanceRegistry.AllInstances() 
	
	insData := []interface{}{}
	
	for _, ins := range instances {
		d := make(map[string]interface{})
		d["start_message"] = ins.StartMessage
		d["container_info"] = ins.ContainerInfo
		insData = append(insData,d)
		
	}
	
	data["save_time"] = nowTime
	data["total_instance_count"] = len(insData)
	data["instances"] = insData
	
	err := util.SaveSnapshot(data, s.path())
	if err != nil {
		s.logger.Errorf("Snapshot save fail,%s", err)
	}
	
	s.logger.Infof("Snapshot save success")
}


func (s *Snapshot) Load() {
	s.logger.Info("Snapshot Begin Load")
	
	snapshotPath := s.path()
	
	_, err := os.Stat(snapshotPath)
	if err != nil {
		s.logger.Errorf("Snapshot Load fail path fail, %s", err.Error() )
		return
	}
	
	bytes , err := ioutil.ReadFile(snapshotPath)
	if err != nil {
		s.logger.Errorf("Snapshot Load readFile fail, %s", err.Error() )
		return
	}
	
	var instances Instances
	err = json.Unmarshal(bytes, &instances)
	
	if err != nil {
		s.logger.Errorf("Snapshot Load Unmarshal fail, %s", err.Error() )
		return
	}
	
	if &instances == nil {
		s.logger.Errorf("Snapshot Load Unmarshal fail instances nil")
		return
	}
	
	if len(instances.Instances) <= 0 {
		s.logger.Errorf("Snapshot Load instances is empty")
		return
	}
	
	for _, instanceAttr := range instances.Instances {
		if instanceAttr.StartMessage == nil {
			continue	
		}
		
		if instanceAttr.ContainerInfo == nil {
			continue	
		}
		
		instance,err := s.instanceManager.CreateInstance(instanceAttr.StartMessage)
		if err != nil {
			s.logger.Errorf("Snapshot Create Instance fail, %s", err.Error() )
			continue
		}
		
		instance.SetState(STATE_RESUMING)
		
		err = instance.startingCompletion(instanceAttr.ContainerInfo.Id)
		if err != nil {
			instance.ContainerInfo = &ContainerInfo{}
			instance.logger.Errorf("Instance start fail,%s", err.Error() )
			//remove image
			instance.docker.RemoveImage(instance.StartMessage.ImageName, instance.StartMessage.ImageTarg)
			//remove container
			instance.docker.RemoveContainer(instanceAttr.ContainerInfo.Id)
			
			instance.SetState(STATE_CRASHED)	
			return
		}
		instance.SetState(STATE_RUNNING)
		
	}
	
}


func (s *Snapshot) path() string {
	return path.Join(s.config.Dea.SnapshotPath, snapshotName) 
}