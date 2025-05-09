package main

import (
	"Go-Chat/api"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

//todo : 连接mongo
//todo : 消息持久化
//todo : api接口提供
//todo : redis维护在线sessions
//todo : kafka动态消费

func main() {
	//var module string
	//flag.StringVar(&module, "module", "", "assign run module")
	//flag.Parse()
	//fmt.Println(fmt.Sprintf("start run %s module", module))
	//
	//switch module {
	////case "logic":
	////	logic.New().Run()
	//case "connect_websocket":
	//	connect.New().Run()
	////case "connect_tcp":
	////	connect.New().RunTcp()
	////case "task":
	////	task.New().Run()
	////case "api":
	////	api.New().Run()
	////case "site":
	////	site.New().Run()
	//default:
	//	fmt.Println("exiting,module param error!")
	//}

	api.New().Run()
	//connect.New().Run()
	//fmt.Println(fmt.Sprintf("run %s module done!", module))
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("Server exiting")
}
