package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
	"time"
)

var (
	once sync.Once
	Conf *AllConfig
)

type MongoConfig struct {
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Host           string `mapstructure:"host"`
	Port           string `mapstructure:"port"`
	DBName         string `mapstructure:"dbname"`
	CollectionName string `mapstructure:"collection_name"`
}
type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Config   string `mapstructure:"config"`
}
type EtcdConfig struct {
	Host              string `mapstructure:"host"`
	BasePath          string `mapstructure:"basePath"`
	ServerPathLogic   string `mapstructure:"serverPathLogic"`
	ServerPathConnect string `mapstructure:"serverPathConnect"`
	UserName          string `mapstructure:"userName"`
	Password          string `mapstructure:"password"`
	ConnectionTimeout int    `mapstructure:"connectionTimeout"`
}
type RedisConfig struct {
	RedisAddress  string `mapstructure:"redisAddress"`
	RedisPassword string `mapstructure:"redisPassword"`
	Db            int    `mapstructure:"db"`
}
type KafkaConfig struct {
	MessageMode string        `mapstructure:"messageMode"`
	HostPort    string        `mapstructure:"hostPort"`
	LoginTopic  string        `mapstructure:"loginTopic"`
	LogoutTopic string        `mapstructure:"logoutTopic"`
	ChatTopic   string        `mapstructure:"chatTopic"`
	Partition   int           `mapstructure:"partition"`
	Timeout     time.Duration `mapstructure:"timeout"`
}
type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	TTL    string `mapstructure:"ttl"`
	Name   string `mapstructure:"name"`
}

type ConnectBase struct {
	CertPath string `mapstructure:"certPath"`
	KeyPath  string `mapstructure:"keyPath"`
}
type ConnectRpcAddressWebsockts struct {
	Address string `mapstructure:"address"`
}
type ConnectRpcAddressTcp struct {
	Address string `mapstructure:"address"`
}
type ConnectBucket struct {
	CpuNum        int    `mapstructure:"cpuNum"`
	Channel       int    `mapstructure:"channel"`
	Room          int    `mapstructure:"room"`
	SrvProto      int    `mapstructure:"svrProto"`
	RoutineAmount uint64 `mapstructure:"routineAmount"`
	RoutineSize   int    `mapstructure:"routineSize"`
}
type ConnectWebsocket struct {
	ServerId string `mapstructure:"serverId"`
	Bind     string `mapstructure:"bind"`
}
type ConnectConfig struct {
	ConnectBase                ConnectBase                `mapstructure:"connect-base"`
	ConnectRpcAddressWebSockts ConnectRpcAddressWebsockts `mapstructure:"connect-rpcAddress-websockts"`
	ConnectRpcAddressTcp       ConnectRpcAddressTcp       `mapstructure:"connect-rpcAddress-tcp"`
	ConnectBucket              ConnectBucket              `mapstructure:"connect-bucket"`
	ConnectWebsocket           ConnectWebsocket           `mapstructure:"connect-websocket"`
}
type LogicConfig struct {
	ServerId   string `mapstructure:"serverId"`
	CpuNum     int    `mapstructure:"cpuNum"`
	RpcAddress string `mapstructure:"rpcAddress"`
	CertPath   string `mapstructure:"certPath"`
	KeyPath    string `mapstructure:"keyPath"`
}
type ApiConfig struct {
	ListenPort int `mapstructure:"listenPort"`
}

type AllConfig struct {
	Logic   LogicConfig   `mapstructure:"logic"`
	Connect ConnectConfig `mapstructure:"connect"`
	Api     ApiConfig     `mapstructure:"api"`

	Mysql MysqlConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
	Etcd  EtcdConfig  `mapstructure:"etcd"`
	Kafka KafkaConfig `mapsturcture:"kafka"`
	Mongo MongoConfig `mapsturcture:"mongo"`
	Jwt   JwtConfig   `mapsturcture:"jwt"`
}

func init() {
	InitConfig()
}
func InitConfig() {
	once.Do(func() {
		config := viper.New()
		config.AddConfigPath("./config")
		config.SetConfigName("application-dev")
		config.SetConfigType("yaml")
		var configData *AllConfig
		err := config.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Use Viper ReadInConfig Fatal error config err:%s \n", err))
		}
		err = config.Unmarshal(&configData)
		if err != nil {
			panic(fmt.Errorf("Use Viper Unmarshal Fatal error config err:%s \n", err))
		}
		fmt.Printf("配置文件信息：%+v\n", configData) // 打印配置文件信息
		Conf = configData
	})
}
