package starting

import (
	"dea-docker/src/dea/dockerapi"
	"strings"
	"errors"
	"fmt"
	"strconv"
)

const (
		Min_Memory = 256//min memory >=256
		Min_Disk   = 50 // min disk >=50
	)

func NewCreateContainerOpts(msg *NatsStartMsg , portpoll *PortPool, instanceId string) (*dockerapi.Container, error ) {
	
	var memory , disk int
	//check limit
	if msg.Limits["mem"] < Min_Memory {
		memory = Min_Memory
	}else {
		memory = msg.Limits["mem"]
	}
	
	if msg.Limits["disk"] < Min_Memory {
		disk = Min_Disk
	}else {
		disk = msg.Limits["disk"]
	} 
	
	//check port
	
	port, err := portpoll.GetPort()
	if err != nil {
		return nil, err	
	}
	
	env, err := getEnv(msg, memory, disk)
	if err != nil {
		return nil,err	
	}
	
	bindPort, port := makeBindport(port, msg, instanceId)
	//registry port to portpool
	portpoll.RegistryPort(port)
	
	return &dockerapi.Container{
		Name:				instanceId,
		Image:				fmt.Sprintf("%s:%s", msg.ImageName, msg.ImageTarg),
		Env:				env,
		Ports:				[]dockerapi.Port{bindPort},
		Memory:				memory,
	},nil
}

func makeBindport(port *Port, msg *NatsStartMsg, instanceId string) (dockerapi.Port, *Port) {
	
	bindPort := dockerapi.Port{
		Name:			"hostport",
		HostPort:		port.Port,
		ContainerPort:	8081,
		Protocol:		"tcp",
		HostIP:			"0.0.0.0",
	}
	
	port.ApplicationId = msg.Droplet
	port.Protocol	= "tcp"
	port.InstanceId = instanceId
	return bindPort, port
}

func getEnv (msg *NatsStartMsg, memory int , disk int)  ([]dockerapi.EnvVar,error) {
	
	if strings.Contains(strings.ToUpper(msg.BuildPack), "JAVA") {
		return makeJavaEnv(msg, memory, disk)	
	}
	
	return nil,errors.New(fmt.Sprintf("getEnv but Application.BuildPack:{%s} not found,", msg.BuildPack))
}

func makeJavaEnv (msg *NatsStartMsg , memlimit int, disklimit int) ([]dockerapi.EnvVar,error) {
	
	env, err := makeEnv(msg.Env)
	if err != nil {
		return nil,err	
	}
	env = append(env, dockerapi.EnvVar{Name: "XMS", Value: fmt.Sprintf("-Xms%sm", strconv.Itoa(memlimit) ), })
	env = append(env, dockerapi.EnvVar{Name: "XMX", Value: fmt.Sprintf("-Xmx%sm", strconv.Itoa(memlimit) ), })
	env = append(env, dockerapi.EnvVar{Name: "MAXPERMISIZE", Value: "-XX:MaxPermSize=256m", })
	
	return env,nil
}

func makeEnv(env []string) ([]dockerapi.EnvVar,error) {
	
	if env == nil || len(env) == 0 {
		return nil, nil	
	}
	
	envVar := []dockerapi.EnvVar{}
	
	for _,s := range env {
		if len(s) != 0 {
			envarr := strings.FieldsFunc(s, func(c rune) bool {return c == '='} )
			if envarr == nil || len(envarr) != 2 {
				return nil,errors.New(fmt.Sprintf("makeEvn fail, split env:%s error ", s ) )				
			}
		
			envVar = append(envVar, dockerapi.EnvVar{Name: envarr[0], Value: envarr[1], })	
		}
	}
	
	return envVar,nil
}