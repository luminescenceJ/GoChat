package logic

import (
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/proto"
	"Go-Chat/tools"
	"context"
	"errors"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"github.com/rpcxio/rpcx-etcd/serverplugin"
	"github.com/sirupsen/logrus"
	"github.com/smallnest/rpcx/server"
	"time"
)

func (logic *Logic) createRpcServer(network string, addr string) {
	s := server.NewServer()
	logic.addRegistryPlugin(s, network, addr)
	// serverId must be unique
	err := s.RegisterName(config.Conf.Etcd.ServerPathLogic, new(RpcLogic), fmt.Sprintf("%s", logic.ServerId))
	if err != nil {
		logrus.Errorf("register error:%s", err.Error())
	}
	s.RegisterOnShutdown(func(s *server.Server) {
		err := s.UnregisterAll()
		if err != nil {
			return
		}
	})
	err = s.Serve(network, addr)
	if err != nil {
		return
	}
}

func (logic *Logic) addRegistryPlugin(s *server.Server, network string, addr string) {
	r := &serverplugin.EtcdV3RegisterPlugin{
		ServiceAddress: network + "@" + addr,
		EtcdServers:    []string{config.Conf.Etcd.Host},
		BasePath:       config.Conf.Etcd.BasePath,
		Metrics:        metrics.NewRegistry(),
		UpdateInterval: time.Minute,
	}
	err := r.Start()
	if err != nil {
		logrus.Fatal(err)
	}
	s.Plugins.Add(r)
}

// 提供外部rpc接口
type RpcLogic struct {
}

func (logic *RpcLogic) Register(ctx context.Context, args *proto.RegisterRequest, reply *proto.RegisterReply) (err error) {
	reply.Code = e.FailReplyCode

	u := new(User)
	uData := u.CheckHaveUserName(args.Name)
	if uData.Id > 0 {
		return errors.New("this user name already have , please login !!!")
	}

	u.UserName = args.Name
	u.Password = args.Password
	u.Uuid = "U" + tools.GetNowAndLenRandomString(11)

	// 存储到mysql
	userId, err := u.Add()
	if err != nil {
		logrus.Infof("register err:%s", err.Error())
		return err
	}
	if userId == 0 {
		return errors.New("register userId empty!")
	}

	//set jwt token
	jwtToken, err := tools.GenerateToken(uint64(userId), config.Conf.Jwt.Name, config.Conf.Jwt.Secret)
	if err != nil {
		logrus.Infof("GenerateToken err:%s", err.Error())
	}

	// redis记录在线人数
	err = SetUserOnline(u.Uuid, 60*time.Second) // 每 30 秒续命心跳
	if err != nil {
		logrus.Infof("SetUserOnline err:%s", err.Error())
	}

	reply.Code = e.SuccessReplyCode
	reply.AuthToken = jwtToken
	return
}

func (logic *RpcLogic) Login(ctx context.Context, args *proto.LoginRequest, reply *proto.LoginResponse) (err error) {
	reply.Code = e.FailReplyCode
	u := new(User)
	userName := args.Name
	passWord := args.Password
	data := u.CheckHaveUserName(userName)
	if (data.Id == 0) || (passWord != data.Password) {
		return errors.New("no this user or password error!")
	}

	//set jwt token
	jwtToken, err := tools.GenerateToken(uint64(data.Id), config.Conf.Jwt.Name, config.Conf.Jwt.Secret)
	if err != nil {
		logrus.Infof("GenerateToken err:%s", err.Error())
	}

	// redis记录在线人数
	err = SetUserOnline(u.Uuid, 60*time.Second) // 每 30 秒续命心跳
	if err != nil {
		logrus.Infof("SetUserOnline err:%s", err.Error())
	}

	reply.Code = e.SuccessReplyCode
	reply.AuthToken = jwtToken
	return
}

func (logic *RpcLogic) AddRoom(ctx context.Context) {

}
