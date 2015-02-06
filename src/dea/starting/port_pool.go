package starting

import (
	"sync"
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/config"
	"net"
)

const(
		PORT_STATE_USED = "USED"
		PORT_STATE_FREE = "FREE"
		PROTOCOL		= "tcp"
		PORT_DEFAULT_START = 61000
		PORT_DEFAULT_END = 69999
	)

type PortPool struct {
	logger					*steno.Logger
	freeLock 				sync.Mutex
	usedLock				sync.Mutex
	ports					[]int
	usedPorts				map[string]*Port
}

type Port struct {
	InstanceId		string
	ApplicationId	string
	ContainerId		string
	Port			int
	Protocol		string
}

func NewPortPool(conf *config.Config) (*PortPool,error){
	
	var portStart, portEnd int
	//init port pool
	if conf.Dea.PortPoolStart == 0 || conf.Dea.PortPoolStart < PORT_DEFAULT_START {
		portStart = PORT_DEFAULT_START
	}else {
		portStart = conf.Dea.PortPoolStart	
	}
	
	if conf.Dea.PortPoolEnd == 0 || conf.Dea.PortPoolEnd < PORT_DEFAULT_END {
		portEnd = PORT_DEFAULT_END
	}else {
		portEnd = conf.Dea.PortPoolEnd
	}
	
	ports := []int{}
	for i:= portStart ; i<= portEnd ; i++ {
		ports= append(ports, i)
	}
	
	return &PortPool{
		logger:		steno.NewLogger("dea-docker"),
		ports:		ports,
		usedPorts:	make(map[string]*Port),
	}, nil
}

//return free port
func (p *PortPool) GetPort() (*Port,error) {
		
	var freePort int 
	for {
		freePort = p.getFreePort()
		if freePort != 0 {
			break	
		}
	}

	port := &Port{
			Port:		freePort,
			Protocol:	"tcp",
		}
	
	return port,nil
}

//registry port to portpool
func (p *PortPool) RegistryPort (port *Port) {
	p.usedLock.Lock()
	defer p.usedLock.Unlock()
	
	p.usedPorts[string(port.Port)] = port
}

//unregistry port to portpool
func (p *PortPool) UnRegistryPort (port *Port) {
	p.usedLock.Lock()
	defer p.usedLock.Unlock()
	
	delete(p.usedPorts, string(port.Port))
	
	p.freePort(port.Port)
}

//release port from portpool
func (p *PortPool) freePort (port int) {
	p.freeLock.Lock()
	defer p.freeLock.Unlock()
	
	p.ports = append(p.ports, port)
}


func (p *PortPool) getFreePort() int {
	p.freeLock.Lock()
	defer p.freeLock.Unlock()
	
	port := p.ports[0]
	
	if checkPort := p.checkPort(port) ; checkPort == false {
		return port	
	}else {
		p.logger.Warnf("PortPool getFreePort but port:{%s} is alredy used", port)
		p.removePortFromPorts()
	}
	
	return 0
}

//remove index=0 from ports
func (p *PortPool) removePortFromPorts() {
	p.ports = append(p.ports[:0], p.ports[1:]...)
}

//if port is already used return true else return false
func (p *PortPool) checkPort(port int) bool {
	
	conn,err := net.DialTCP("tcp", nil, &net.TCPAddr{IP:net.ParseIP("127.0.0.1"), Port:port,})
	
	if err != nil {
		return false	
	}
	
	err = conn.Close()
	if err != nil {
		p.logger.Warnf("CheckPort and Close conn fail, %s", err)	
	}
	return true
}
