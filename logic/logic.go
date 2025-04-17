package logic

import (
	"Go-Chat/config"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"runtime"
)

type Logic struct {
	ServerId string
}

func New() *Logic {
	return new(Logic)
}

func (logic *Logic) Run() {
	//read config
	logicConfig := config.Conf.Logic

	runtime.GOMAXPROCS(logicConfig.CpuNum)
	logic.ServerId = fmt.Sprintf("logic-%s", uuid.New().String())

	//init publish redis
	if err := logic.InitRedisClient(); err != nil {
		logrus.Panicf("logic InitRedisClient fail,err:%s", err.Error())
		return
	}

	//init rpc server
	if err := logic.InitRpcServer(); err != nil {
		logrus.Panicf("logic init rpc server fail")
		return
	}
}
