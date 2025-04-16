package connect

import (
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/proto"
	"Go-Chat/tools"
	"context"
	"github.com/go-redis/redis"
	"github.com/rpcxio/libkv/store"
	etcd "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"strings"
	"sync"
	"time"
)

var (
	RedisClient    *redis.Client
	KafkaClient    *tools.KafkaService
	logicRpcClient client.XClient
	once           sync.Once
)

type RpcConnect struct{}

func (rpc *RpcConnect) Connect(connReq *proto.ConnectRequest) (uid int, err error) {
	reply := &proto.ConnectReply{}
	err = logicRpcClient.Call(context.Background(), "Connect", connReq, reply)
	if err != nil {
		logrus.Fatalf("failed to call: %v", err)
	}
	uid = reply.UserId
	logrus.Infof("connect logic userId :%d", reply.UserId)
	return
}

func (rpc *RpcConnect) DisConnect(disConnReq *proto.DisConnectRequest) (err error) {
	reply := &proto.DisConnectReply{}
	if err = logicRpcClient.Call(context.Background(), "DisConnect", disConnReq, reply); err != nil {
		logrus.Fatalf("failed to call: %v", err)
	}
	return
}

func (c *Connect) InitLogicRpcClient() (err error) {
	etcdConfigOption := &store.Config{
		ClientTLS:         nil,
		TLS:               nil,
		ConnectionTimeout: time.Duration(config.Conf.Etcd.ConnectionTimeout) * time.Second,
		Bucket:            "",
		PersistConnection: true,
		Username:          config.Conf.Etcd.UserName,
		Password:          config.Conf.Etcd.Password,
	}
	once.Do(func() {
		d, err := etcd.NewEtcdV3Discovery(
			config.Conf.Etcd.BasePath,        // 基础路径
			config.Conf.Etcd.ServerPathLogic, // 服务子路径
			[]string{config.Conf.Etcd.Host},  // etcd 集群地址
			true,                             // 是否监听服务变化
			etcdConfigOption,                 // 高级配置（超时、认证等）
		)
		if err != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail:%s", err.Error())
		}

		logicRpcClient = client.NewXClient(config.Conf.Etcd.ServerPathLogic,
			client.Failtry,      // 启用失败重试
			client.RandomSelect, // 随机选择服务实例
			d, client.DefaultOption)
	})
	if logicRpcClient == nil {
		return e.Error_RPC_CREATE
	}
	return err
}

func (c *Connect) InitKafkaClient() (err error) {
	KafkaClient = tools.GetKafkaInstance(config.Conf.Kafka)
	return err
}

func (c *Connect) InitRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Redis.RedisAddress,
		Password: config.Conf.Redis.RedisPassword,
		Db:       config.Conf.Redis.Db,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisCli Ping Result pong: %s,  err: %s", pong, err)
	}
	return err
}

func (c *Connect) InitConnectWebsocketRpcServer() (err error) {
	var network, addr string
	connectRpcAddress := strings.Split(config.Conf.Connect.ConnectRpcAddressWebSockts.Address, ",")
	for _, bind := range connectRpcAddress {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitConnectWebsocketRpcServer ParseNetwork error : %s", err)
		}
		logrus.Infof("Connect start run at-->%s:%s", network, addr)
		go c.createConnectWebsocktsRpcServer(network, addr)
	}
	return err
}
