package connect

import (
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/proto"
	"context"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"time"
)

type RpcConnectPush struct{}

// 私聊消息rpc
func (rpc *RpcConnectPush) PrivateMsg(ctx context.Context, msg *proto.Message, successReply *proto.SuccessReply) (err error) {
	var (
		bucket *Bucket
		s      *Session
	)
	logrus.Info("rpc Private message :%v ", msg)
	if msg == nil {
		logrus.Errorf("rpc PushSingleMsg() args:(%v)", msg)
		return
	}
	bucket = DefaultServer.Bucket(msg.SenderId)
	if s = bucket.GetSession(msg.SenderId); s != nil {
		err = s.SendMsg(msg)
		logrus.Infof("DefaultServer Channel err nil ,args: %v", msg)
		return
	}
	successReply.Code = e.SuccessReplyCode
	successReply.Msg = e.SuccessReplyMsg
	logrus.Infof("successReply:%v", successReply)
	return
}

// 群发消息rpc
func (rpc *RpcConnectPush) GroupMsg(ctx context.Context, msg *proto.Message, successReply *proto.SuccessReply) (err error) {
	successReply.Code = e.SuccessReplyCode
	successReply.Msg = e.SuccessReplyMsg
	logrus.Infof("rpc Group message :  %+v", msg)
	for _, bucket := range DefaultServer.Buckets {
		bucket.BroadcastRoom(msg)
	}
	return
}

func (c *Connect) createConnectWebsocktsRpcServer(network string, addr string) {
	// 初始化服务器
	s := server.NewServer()

	// 注册中心插件初始化
	if err := addRegistryPlugin(s, network, addr); err != nil {
		logrus.Errorf("Failed to add registry plugin: %v", err)
		return
	}

	// 注册 RPC 服务
	if err := s.RegisterName(
		config.Conf.Etcd.ServerPathConnect,
		new(RpcConnectPush),
		fmt.Sprintf("serverId=%s&serverType=ws", c.ServerId),
	); err != nil {
		logrus.Errorf("Failed to register RPC service: %v", err)
		return
	}

	// 优雅关闭处理
	shutdownErrChan := make(chan error, 1)
	s.RegisterOnShutdown(func(s *server.Server) {
		logrus.Info("Starting graceful shutdown...")
		if err := s.UnregisterAll(); err != nil {
			shutdownErrChan <- fmt.Errorf("failed to unregister services: %w", err)
			return
		}
		shutdownErrChan <- nil
	})

	// 启动服务
	logrus.Infof("Starting RPC server on @%s://%s", network, addr)
	if err := s.Serve(network, addr); err != nil {
		logrus.Errorf("RPC server failed: %v", err)

		// 强制关闭处理
		if shutdownErr := s.Close(); shutdownErr != nil {
			logrus.Errorf("Force shutdown error: %v", shutdownErr)
		}
		return
	}

	// 等待关闭结果
	if err := <-shutdownErrChan; err != nil {
		logrus.Errorf("Shutdown completed with errors: %v", err)
	} else {
		logrus.Info("Shutdown completed successfully")
	}
}

func addRegistryPlugin(s *server.Server, network string, addr string) error {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr, // 示例："tcp@192.168.1.100:8090"
		EtcdServers:    []string{config.Conf.Etcd.Host},
		BasePath:       config.Conf.Etcd.BasePath, // 如 "/gochat_srv"
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute, // 心跳间隔
	}

	if err := r.Start(); err != nil {
		return fmt.Errorf("etcd plugin start failed: %w", err)
	}

	s.Plugins.Add(r)
	return nil
}
