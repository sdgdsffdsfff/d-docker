package main 

import (
	"github.com/cloudfoundry/yagnats"
 	steno "github.com/cloudfoundry/gosteno"
 	"dea-docker/src/dea/config"
 	codec "dea-docker/src/dea/logger"
 	"flag"
 	"strings"
 	"fmt"
 	"time"
 	"os"
 	"dea-docker/src/dea/controller"
	"dea-docker/src/dea/util"
	"dea-docker/src/dea/api"
	"dea-docker/src/dea/starting"
)

//定义基础变量
var (
	natsClient *yagnats.Client
	logger     		*steno.Logger
	configFile 		string
	conf		 		*config.Config
	ctrl		 		*controller.Controller
	nats		 		*api.Nats
	natsHandle	 		*api.NatsMessageHandle
	instanceManage	*starting.InstanceManager
	instancesRegistry *starting.InstanceRegistry
	resourceManager	*starting.ResourceManager
	snapshot			*starting.Snapshot
	localIp 			string
	uuid				string
)

//初始化配置信息
func setupConfig () {

	fmt.Printf("setup config filePath:%s \n",configFile)
	
	conf = config.DefaultConfig()
	
	if configFile != "" {
		conf = config.InitConfigFromFile(configFile);
	}
}

//初始化Nats信息
func setupNats () {
	
	fmt.Printf("setup nats ,host:%s, port:%s,user:%s,pass:%s \n",conf.Nats.Host,conf.Nats.Port,conf.Nats.User,conf.Nats.Pass)
	
	var err error
	natsClient = yagnats.NewClient()
	addr := conf.Nats.Host
	
	if !strings.HasPrefix(addr, "zk://") {
		addr = fmt.Sprintf("%s:%d", conf.Nats.Host, conf.Nats.Port)
	}
	
	natsInfo := &yagnats.ConnectionInfo{
		Addr:     addr,
		Username: conf.Nats.User,
		Password: conf.Nats.Pass,
	}
	
	err = natsClient.Connect(natsInfo)
	
	for ; err != nil; {
		err = natsClient.Connect(natsInfo)
		fmt.Printf("natsClient.connect fial, %s",err)
		time.Sleep(500 * time.Millisecond)
	}
	
}

//初始化日志信息
func setupLogger () {

	fmt.Printf("setup logger level:%s,file:%s \n",conf.Logging.Level,conf.Logging.File)
	
	l, err := steno.GetLogLevel(conf.Logging.Level)
	if err != nil {
		logger.Errorf("steno.GetLogLevel fail , %s", err)
		os.Exit(1)
	}

	s := make([]steno.Sink, 0, 3)
	s = append(s, steno.NewFileSink(conf.Logging.File))

	stenoConfig := &steno.Config{
		Sinks: s,
		Codec: codec.NewStringCodec(),
		Level: l,
	}

	steno.Init(stenoConfig)
	
	logger = steno.NewLogger("dea-docker")
}

//初始化http客户端
func setupHttp () {
	ctrl = controller.NewController(conf, instancesRegistry)
}

//初始化基本信息
func setupInfo () {

	var err error
	
	//local ip
	localIp, err = util.GetLocalIp()
	
	if err != nil {
		logger.Errorf("get local ip fail,%s", err)
		os.Exit(1)
	}
	
	//uuid
	var guid string
	guid , err = util.GetGuid()
	
	if err != nil {
		logger.Errorf("get guid fail, %s", guid)
		os.Exit(1)
	}
	
//	uuid = conf.Dea.Index +"-"+guid
	uuid = "test"
	conf.Dea.LocalIp = localIp
	conf.Dea.Uuid	=  uuid
	fmt.Printf("localIp:%s \n",localIp)
	fmt.Printf("uuid : %s \n", uuid)
}

func setupSnapshot () {
	snapshot = starting.NewSnapshot(conf)
}
func setupResourceManager () {
	resourceManager = starting.NewResourceManager(conf)
}

func setupInstanceManager () {
	
	var err error
	instanceManage,err = starting.NewInstanceManager(natsClient, resourceManager, instancesRegistry, snapshot, conf);
	if err != nil {
		logger.Errorf("create InstanceManager fail,%s", err)
		os.Exit(1)	
	}
}

func setupInstancesRegistry () {
	instancesRegistry = starting.NewInstanceRegistry(natsClient, conf)
}

func setupNatsHandle () {
	natsHandle = api.NewNatsMessageHandle(instanceManage, instancesRegistry)
	nats = api.NewNats(natsClient, natsHandle,uuid)
}

func configure() {
	snapshot.Configure(instancesRegistry, instanceManage)
}

func setup ()	{
	setupConfig()
	setupInfo()
	setupLogger()
	setupNats()
	setupSnapshot()
	setupInstancesRegistry()
	setupResourceManager()
	setupInstanceManager()
	
	setupNatsHandle()
	setupHttp()
}

func start () {
	
	//start nats
	nats.Start()
	//start instanceRegistry
	instancesRegistry.Start()
	//snapshot load
	snapshot.Load()
	//start server api
	ctrl.ServeApi()
	
	fmt.Println("starting success")
}
//解析启动参数,-c 配置文件路径
func init(){
	flag.StringVar(&configFile, "c", "", "Configuration File")
	
	flag.Parse()
}

func main() {
	setup()
	configure()
	start()
}

