package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
)

type AppConfig struct {
	Server     Server     `yaml:"server"`
	Etcd       Etcd       `yaml:"etcd"`
	Lustre     Lustre     `yaml:"lustre"`
	Controller Controller `yaml:"controller"`
}

type Server struct {
	Port string `yaml:"port"`
	IP   string `yaml:"ip"`
}

type Etcd struct {
	Endpoints   []string `yaml:"endpoints"`
	Dialtimeout int      `yaml:"dial_timeout"`
	Leasettl    int      `yaml:"lease_ttl"`
}

type Lustre struct {
	Mkfsoptions string `yaml:"mkfsoptions"`
	Backfstype  string `yaml:"backfstype"`
	Mgs         string `yaml:"mgs"`
	Mdt         string `yaml:"mdt"`
	Common      string `yaml:"common"`
}

type Controller struct {
	Name string `yaml:"name"`
	Node Node   `yaml:"node"`
}
type Node struct {
	A string `yaml:"A"`
	B string `yaml:"B"`
}

var (
	Config     *viper.Viper
	ConfigData AppConfig
)

func init() {
	log.Println("Loading configuration logics...")
	Config = initConfig()
	go dynamicReloadConfig()
}

func initConfig() *viper.Viper {
	Config = viper.New()
	Config.SetConfigName("app")
	Config.SetConfigType("yaml")
	Config.AddConfigPath("./config/conf/")
	err := Config.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	//查找并读取配置文件
	err = Config.ReadInConfig()
	if err != nil { // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if err := Config.Unmarshal(&ConfigData); err != nil { // 读取配置文件转化成对应的结构体错误
		panic(fmt.Errorf("read config file to struct err: %s \n", err))
	}

	return Config
}

func dynamicReloadConfig() {
	Config.WatchConfig()
	Config.OnConfigChange(func(event fsnotify.Event) {
		log.Printf("Detect config change: %s \n", event.String())
	})
}
