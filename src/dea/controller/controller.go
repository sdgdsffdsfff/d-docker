package controller

import (
	"net/http"
 	"github.com/gorilla/mux"
	"dea-docker/src/dea/config"
	steno "github.com/cloudfoundry/gosteno"
	"dea-docker/src/dea/starting"
	"net"
	"fmt"
	"os"
)

// 提供restfull 接口
type Controller struct {
	conf 				*config.Config 
	logger     			*steno.Logger
	instanceRegister 	*starting.InstanceRegistry
}

type HttpApiFunc func(w http.ResponseWriter, r *http.Request, vars map[string]string) error

func NewController (c *config.Config ,insRegister *starting.InstanceRegistry) *Controller {
	return &Controller {
		conf:				c,
		logger: 			steno.NewLogger("dea-docker"),
		instanceRegister	:insRegister,
	}
}

func (c *Controller) returnJson(v interface{}, w http.ResponseWriter) error{
	data, err := encodeJson(v)
	
	if err != nil {
		c.logger.Errorf("encodejson fail,err:%v",err)
		return err
	}else {
		writeJson(data, w)
	}
	return nil
}

func (c *Controller) configInfo(w http.ResponseWriter, r *http.Request, vars map[string]string ) error {
	err := c.returnJson(c.conf, w)
	return err
}

func (c *Controller) containerinfo(w http.ResponseWriter, r *http.Request, vars map[string]string ) error {
	
	instances := c.instanceRegister.AllInstances()
	err := c.returnJson(instances, w)
	
	return err	
}

func (c *Controller) containerFilter(w http.ResponseWriter, r *http.Request, vars map[string]string ) error {

	applicationId, ok := vars["applicationId"]
	
	if !ok {
		writeStr("applicationId is empty", w)
		return nil
	}
	
	err := c.returnJson(c.instanceRegister.FilterInstanceByApplication(applicationId) , w)
	
	return err
}


func (c *Controller) makeHttpHandler(logging bool, localMethod string, localRouter string, handlerFunc HttpApiFunc) http.HandlerFunc {
	
	return func(w http.ResponseWriter, r *http.Request) {
		c.logger.Infof("Calling %s %s", localMethod, localRouter)
		
		if logging {
			c.logger.Infof("reqMethod:%s , reqURI:%s , userAgent:%s", r.Method, r.RequestURI, r.Header.Get("User-Agent"))
		}
		
		if err := handlerFunc(w , r, mux.Vars(r)) ; err != nil {
			c.logger.Errorf("Handler for %s %s returned error: %s", localMethod, localRouter, err)
			http.Error(w, err.Error(), 400)
		}
	}
}

func (c *Controller) createoRuter () (*mux.Router, error) {
	r := mux.NewRouter()
	
	m := map[string]map[string] HttpApiFunc {
		"GET": {
			"/configinfo":							c.configInfo,
			"/container":							c.containerinfo,
			"/container/{applicationId:.*}/get":	c.containerFilter,
		},
		"POST": {
		
		},
		"DELETE": {
		
		},
		"PUT": {
		
		},
	}
	
	//遍历定义的方法,注册服务
	for method, routers := range m {
		
		for route, fct := range routers {
			c.logger.Infof("registering method:%s, router:%s", method, route)
			
			localRoute := route
			localFct   := fct
			localMethod := method
			
			//build the handler function
			f := c.makeHttpHandler(c.conf.Dea.HandlerLogging, localMethod, localRoute, localFct)
			
			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			}else {
				r.Path("/" + c.conf.Dea.Basepath + localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}
	
	return r, nil
}

// 开启服务监听
func (c *Controller) listenAndServe() error {
	
	var l net.Listener
	r, err := c.createoRuter()
	
	if err != nil {
		return err
	}
	
	addr := ":"+c.conf.Dea.Port
	
	l, err  = net.Listen("tcp", addr)
	if err != nil {
		c.logger.Errorf("listenAndServe fail, %s", err)	
		return err
	}
	httpSrv := http.Server{Addr: addr, Handler: r}
	
	return httpSrv.Serve(l)
}

func (c *Controller) ServeApi() {
	err :=  c.listenAndServe()
	if err != nil {
		fmt.Printf("ServeApi error , %s", err)
		os.Exit(1)
	}
}