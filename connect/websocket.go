package connect

import (
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/tools"
	"context"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
)

// todo:websocket鉴权完成，下一步是服务层
func (c *Connect) InitWebsocket() error {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// 解析jwt
		token := r.URL.Query().Get("token")

		// 前端代码
		// const token = 'your_jwt_token'
		// const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		payLoad, err := tools.ParseToken(token, config.Conf.Jwt.Secret)
		if err != nil {
			logrus.Warnln("parse token fail:", err.Error())
		}

		// ✅ 用 context.WithValue 注入 userId
		ctx := context.WithValue(r.Context(), e.CurrentId, payLoad.UserId)
		r = r.WithContext(ctx)

		c.serveWs(DefaultServer, w, r)

	})
	err := http.ListenAndServe(config.Conf.Connect.ConnectWebsocket.Bind, nil)
	return err
}

func (c *Connect) serveWs(server *Server, w http.ResponseWriter, r *http.Request) {

	var upGrader = websocket.Upgrader{
		ReadBufferSize:  server.Options.ReadBufferSize,
		WriteBufferSize: server.Options.WriteBufferSize,
	}

	//cross origin domain support
	upGrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upGrader.Upgrade(w, r, nil)

	if err != nil {
		logrus.Errorf("serverWs err:%s", err.Error())
		return
	}
	var s *Session
	//default broadcast size eq 512
	s = NewSession(server.Options.BroadcastSize)
	s.conn_ws = conn
	//send data to websocket conn
	go server.writePump(s, c)
	//get data from websocket conn
	go server.readPump(s, c)
}
