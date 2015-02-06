package config

//配置文件
import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"io/ioutil"
)


//nats config
type NatsConfig struct {
	
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
	
}

//默认的Nats配置信息
var defaultNatsConfig = NatsConfig {

	Host: "127.0.0.1",
	Port: 4222,
	User: "nats",
	Pass: "nats",
}

// 日志配置
type LoggingConfig struct {
	File  string `yaml:"file"`
	Level string `yaml:"level"`
}

// 日志配置默认值
var defaultLoggingConfig = LoggingConfig{
	File: "/export/home/jae/dea-docker.log",
	Level: "debug",
}

//dea config
type DeaConfig struct {
	Port				string		 	`yaml:"port"`
	HandlerLogging 		bool 			`yaml:"handlerLogging"`
	Basepath 			string  		`yaml:"basePath"`
	Index	  			string			`yaml:"index"`
	SnapshotPath 		string 			`yaml:"snapshot_path"`
	MemoryMb			int 			`yaml:"memory_mb"`
	DiskMb				int				`yaml:"disk_mb"`
	MemoryFactor		int	 			`yaml:"memory_factor"`
	DiskFactor			int	 			`yaml:"disk_factor"`
	StandbyMemory		int				`yaml:"standby_memory"`
	StandbyDisk			int				`yaml:"standby_disk"`
	DiskPath			string			`yaml:"disk_path"`
	PortPoolStart		int				`yaml:"port_pool_start"`
	PortPoolEnd			int				`yaml:"port_pool_end"`
	LocalIp				string
	Uuid				string
}

var defaultDeaConfig = DeaConfig{
	Port:				"34504",
	HandlerLogging: 	true,
	Basepath:			"dea-docker",
	Index:				"01",
	SnapshotPath:		"/export/data/",
	MemoryMb:			1024,
	DiskMb:				1024,
	MemoryFactor:		32,
	DiskFactor:			200,
	StandbyMemory: 		1024,
	StandbyDisk:		2048,
	PortPoolStart:		61000,
	PortPoolEnd:		69999,
	DiskPath:		"/export/Data",
}

//docker config
type DockerConfig struct {
	Url				string				`yaml:"docker_url"`
	Version			string				`yaml:"docker_version"`
	Registry		string				`yaml:"docker_registry"`
	DockerPath		string				`yaml:"docker_path"`
}

var defaultDockerConfig = DockerConfig{
	Url:					"http://127.0.0.1:4243",
	Version:				"",
	Registry:				"docker.registry.com",
	DockerPath:				"/var/lib/docker/",
}

type Config struct {

	Nats 				 NatsConfig 					`yaml:"nats"`
	Logging			 LoggingConfig              		`yaml:"logging"`
	Dea					 DeaConfig						`yaml:"dea"`
	Docker				 DockerConfig					`yaml:"docker"`
	
}

var defaultConfig	 = Config {
	Nats:		defaultNatsConfig,
	Logging:	defaultLoggingConfig,
	Dea:		defaultDeaConfig,
	Docker:	defaultDockerConfig,
}

func DefaultConfig() *Config{
	
	c := defaultConfig
	
	return &c
}

//解析配置文件,初始化配置对象
func (c *Config) Initialize(configYAML [] byte) error{

	return candiedyaml.Unmarshal(configYAML, &c)
}

//根据文件初始化配置对象
func InitConfigFromFile(path string) *Config{

	var c *Config = DefaultConfig()
	var e error
	
	b, e := ioutil.ReadFile(path)
	
	if e != nil {
		panic(e.Error())
	}
	
	e = c.Initialize(b)
	
	return c
}