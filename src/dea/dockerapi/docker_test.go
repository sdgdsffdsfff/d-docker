package dockerapi

import (
	"testing"
	"fmt"
	"encoding/json"
	"icode.jd.com/cdlxyong/go-dockerclient"
	"dea-docker/src/dea/config"
	"dea-docker/src/dea/dockerapi"
)


func TestListContainers(t *testing.T) {
	client, err  := docker.NewClient("http://127.0.0.1:4243")
	if err != nil {
		t.Errorf("NewClient err,%s",err)
	}
	
	images , err := client.ListImages(true)
	
	if err != nil {
		t.Errorf("ListImages fail , %s",err)	
	}
	
	d , err := json.Marshal(images)
	
	if err != nil {
		t.Errorf("json convert fail ,%s", err)	
	}
	fmt.Printf("------------->result:%s",string(d))
	
}


func TestRunContainer(t *testing.T) {
	config := config.DefaultConfig()
	
	docker , err := dockerapi.NewDocker(config)
	
	if err != nil {
		t.Errorf("--------->fail,%s", err)	
	}
	
	container, err := docker.RunContainer(&dockerapi.Container{
			Name:				"testcontainer3",
			Image:				"java_tomcat6.0.33_jdk1.6.0_25:v2",
			Env:				[]dockerapi.EnvVar{
									dockerapi.EnvVar{Name:"LOG_PATH",Value:"/export/Log/"},
									dockerapi.EnvVar{Name:"JAVA_HOME",Value:"/export/Jdk/jdk1.6.0_25"},
									dockerapi.EnvVar{Name:"JAVA_BIN",Value:"/export/Jdk/jdk1.6.0_25/bin"},
									dockerapi.EnvVar{Name:"PATH",Value:"/export/Jdk/jdk1.6.0_25/bin:/usr/kerberos/sbin:/usr/kerberos/bin:/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin:/root/bin:/bin"},
									dockerapi.EnvVar{Name:"CLASSPATH",Value:".:/lib/dt.jar:/lib/tools.jar"},
									dockerapi.EnvVar{Name:"JAVA_OPTS",Value:"-Djava.library.path=/usr/local/lib -server -Xms2048m -Xmx4000m -XX:MaxPermSize=256m -XX:+HeapDumpOnOutOfMemoryError -XX:ErrorFile=${LOG_PATH}jvm.log -XX:HeapDumpPath=${LOG_PATH}jvm.dump -Djava.awt.headless=true -Dsun.net.client.defaultConnectTimeout=60000 -Dsun.net.client.defaultReadTimeout=60000 -Djmagick.systemclassloader=no -Dnetworkaddress.cache.ttl=300 -Dsun.net.inetaddr.ttl=300"},
									dockerapi.EnvVar{Name:"XMS",Value:"-Xms1024m"},
									dockerapi.EnvVar{Name:"XMX",Value:"-Xmx1024m"},
									dockerapi.EnvVar{Name:"MAXPERMISIZE",Value:"-XX:MaxPermSize=256m"},
								},
			Ports:				[]dockerapi.Port{dockerapi.Port{
									Name:		"hostport",
									HostPort:	8082,
									ContainerPort:	8081,
									Protocol:		"tcp",
									HostIP:			"0.0.0.0",
								}},
			//Command:			[]string{""},
			
		}, "")
	
	if err != nil {
		t.Errorf("runcontainer fail,%s", err)	
	}
	
	d,_ := json.Marshal(container)
	
	fmt.Printf("----------->result:%s", string(d))
		
}