package rpc

import (
	"Go-Chat/config"
	"Go-Chat/proto"
	"context"
	"github.com/rpcxio/libkv/store"
	etcdV3 "github.com/rpcxio/rpcx-etcd/client"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/client"
	"sync"
	"time"
)

var LogicRpcClient client.XClient
var once sync.Once

type RpcLogic struct {
}

var RpcLogicObj *RpcLogic

func (r *RpcLogic) Login(req *proto.LoginRequest) (code int, authToken string, msg string) {
	reply := &proto.LoginResponse{}
	err := LogicRpcClient.Call(context.Background(), "Login", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) Register(req *proto.RegisterRequest) (code int, authToken string, msg string) {
	reply := &proto.RegisterReply{}
	err := LogicRpcClient.Call(context.Background(), "Register", req, reply)
	if err != nil {
		msg = err.Error()
	}
	code = reply.Code
	authToken = reply.AuthToken
	return
}

func (rpc *RpcLogic) Logout(req *proto.LogoutRequest) (code int) {
	reply := &proto.LogoutResponse{}
	err := LogicRpcClient.Call(context.Background(), "Logout", req, reply)
	if err != nil {
		logrus.Error("rpc logout error:", err)
	}
	code = reply.Code
	return
}

func InitLogicRpcClient() {
	once.Do(func() {
		etcdConfigOption := &store.Config{
			ClientTLS:         nil,
			TLS:               nil,
			ConnectionTimeout: time.Duration(config.Conf.Etcd.ConnectionTimeout) * time.Second,
			Bucket:            "",
			PersistConnection: true,
			Username:          config.Conf.Etcd.UserName,
			Password:          config.Conf.Etcd.Password,
		}
		//这个 d 是一个服务发现器，会自动监控 /gochat/logic/ 路径下有哪些服务可用（服务端已注册）。
		d, err := etcdV3.NewEtcdV3Discovery(
			config.Conf.Etcd.BasePath,
			config.Conf.Etcd.ServerPathLogic,
			[]string{config.Conf.Etcd.Host},
			true,
			etcdConfigOption,
		)
		if err != nil {
			logrus.Fatalf("init connect rpc etcd discovery client fail:%s", err.Error())
		}
		// 这一步把服务发现 + 调用策略 + 负载均衡绑定到了 LogicRpcClient 上。
		// 后续通过 LogicRpcClient.Call() 就可以调用远程逻辑服务了。
		LogicRpcClient = client.NewXClient(config.Conf.Etcd.ServerPathLogic, client.Failtry, client.RandomSelect, d, client.DefaultOption)
		RpcLogicObj = new(RpcLogic)
	})
	if LogicRpcClient == nil {
		logrus.Fatalf("get logic rpc client nil")
	}
}
