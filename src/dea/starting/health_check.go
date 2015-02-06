package starting

import (
	steno "github.com/cloudfoundry/gosteno"
	"time"
	"sync"
	"net"
)

type HealthCheck struct {
	
	lock			sync.Mutex
	ticker			*time.Ticker
	logger			*steno.Logger
}

func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		logger:		steno.NewLogger("dea-docker"),
	}
}

func (h *HealthCheck) stop () {
	if h.ticker != nil {
		h.ticker.Stop()
	}
}

func (h *HealthCheck) health(ins *Instance) error{
	h.logger.Infof("Begin HealthCheak,applicationId:%s,InstanceId:%s", ins.ApplicationId, ins.InstanceId)
	h.lock.Lock()
	h.ticker = time.NewTicker(time.Second * 1)
	h.lock.Unlock()
	go func() {
			for {
				select {
				case <-h.ticker.C:
					conn,err := net.DialTCP("tcp", nil, &net.TCPAddr{IP:net.ParseIP(ins.ContainerInfo.HostIp), Port: ins.ContainerInfo.HostPort,})
					
					if err != nil {
						h.logger.Errorf("HealthCheck applicationId:%s,InstanceId:%s , fail:%s ",ins.ApplicationId, ins.InstanceId, err.Error() )
						ins.SetState(STATE_CRASHED)
					}else{
						conn.Close()
					}
				}
			}
		}()
	h.logger.Infof("Success HealthCheak,applicationId:%s,InstanceId:%s", ins.ApplicationId, ins.InstanceId)
	return nil
}
