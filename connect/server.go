package connect

import (
	"Go-Chat/config"
	"Go-Chat/proto"
	"Go-Chat/tools"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

var ctx = context.Background()

type Server struct {
	Buckets   []*Bucket
	Options   ServerOptions
	bucketIdx uint32
	operator  Operator
}

type ServerOptions struct {
	WriteWait       time.Duration
	PongWait        time.Duration
	PingPeriod      time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	BroadcastSize   int
}

func NewServer(b []*Bucket, o Operator, options ServerOptions) *Server {
	s := new(Server)
	s.Buckets = b
	s.Options = options
	s.bucketIdx = uint32(len(b))
	s.operator = o
	return s
}

// reduce lock competition, use google city hash insert to different bucket
func (s *Server) Bucket(userId int) *Bucket {
	userIdStr := fmt.Sprintf("%d", userId)
	idx := tools.CityHash32([]byte(userIdStr), uint32(len(userIdStr))) % s.bucketIdx
	return s.Buckets[idx]
}

//todo : http接收到之后再携带token访问websocket

// 向客户端发送，消费消息和心跳监听
func (s *Server) writePump(ch *Session, c *Connect) {
	// 定时发送心跳
	ticker := time.NewTicker(s.Options.PingPeriod)
	defer func() {
		ticker.Stop()
		ch.conn_ws.Close()
	}()

	for {
		select {
		case msg, ok := <-ch.cache:
			// 写信息超时处理 , default 10s
			if !ok {
				logrus.Warn("SetWriteDeadline not ok")
				ch.conn_ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			ch.conn_ws.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			w, err := ch.conn_ws.NextWriter(websocket.TextMessage)
			if err != nil {
				logrus.Warn(" ch.conn.NextWriter err :%s  ", err.Error())
				return
			}

			msgBytes, err := json.Marshal(msg)
			if err != nil {
				logrus.Warn(" json.Marshal err :%s  ", err.Error())
				return
			}
			write, err := w.Write(msgBytes)
			if err != nil || write != len(msgBytes) {
				logrus.Warn(" write err :%s  ", err.Error())
			}
			// todo : 写入mongo消息状态变化

			logrus.Infof("message write :%s", msgBytes)
			if err = w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// heartbeat，if ping error will exit and close current websocket conn
			ch.conn_ws.SetWriteDeadline(time.Now().Add(s.Options.WriteWait))
			logrus.Infof("websocket.PingMessage :%v", websocket.PingMessage)
			if err := ch.conn_ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

// 从客户端接收，生产消息并持久化
func (s *Server) readPump(ch *Session, c *Connect) {
	// 流程实现：
	// websocket读取前端发送消息类（私聊，群聊，应答，断开连接）
	// 私聊群聊持久化到redis后加入消息队列
	// 消息队列持续消费消息发送到对应的session，再由session进行消费到目标客户端
	defer func() {
		logrus.Infof("start exec disConnect ...")
		//if ch.Room == nil || ch.userId == 0 {
		//	logrus.Infof("roomId and userId eq 0")
		//	ch.conn_ws.Close()
		//	return
		//}
		logrus.Infof("exec disConnect ...")
		//todo : 处理登出逻辑
		//disConnectRequest := new(proto.DisConnectRequest)
		//disConnectRequest.RoomId = ch.Room.Id
		//disConnectRequest.UserId = ch.userId
		//s.Bucket(ch.userId).Delete(ch)
		//if err := s.operator.DisConnect(disConnectRequest); err != nil {
		//	logrus.Warnf("DisConnect err :%s", err.Error())
		//}
		ch.conn_ws.Close()
	}()
	ch.conn_ws.SetReadLimit(s.Options.MaxMessageSize)
	ch.conn_ws.SetReadDeadline(time.Now().Add(s.Options.PongWait))
	ch.conn_ws.SetPongHandler(func(string) error {
		ch.conn_ws.SetReadDeadline(time.Now().Add(s.Options.PongWait))
		return nil
	})
	for {
		_, message, err := ch.conn_ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.Errorf("readPump ReadMessage err:%s", err.Error())
				return
			}
		}
		if message == nil {
			return
		}
		logrus.Infof("get a message :%s", message)
		transMsg := &proto.TransMsg{}
		// 解构消息结构体
		if err = json.Unmarshal(message, transMsg); err != nil {
			logrus.Errorf("transMsg json Umarshal error: %+v", err)
		}

		// 消息基础校验
		if transMsg.SenderId == 0 {
			logrus.Warn("Received message without sender ID")
			continue
		}

		switch transMsg.Type {
		case proto.MsgTypePrivate:

			if transMsg.TargetId == 0 {
				logrus.Warn("Private message missing target ID!")
				continue
			}

			parsedTime, err := time.Parse(time.RFC3339, transMsg.SendTimeStamp)
			if err != nil {
				logrus.Warn("ParseTime err :%s  ", err.Error())
			}

			// 用户在线
			msg := &proto.Message{
				MsgId:         tools.GetSnowflakeId(),
				Type:          transMsg.Type,
				Buffer:        []byte(transMsg.Buffer),
				MsgStatus:     proto.MsgStatusSending,
				RoomId:        transMsg.RoomId,
				SenderId:      transMsg.SenderId,
				TargetId:      transMsg.TargetId,
				SendTimeStamp: parsedTime,
				//ReceiveTimeStamp: time.Time{},
			}

			// redis持久化
			err = c.SaveMsg(msg)
			if err != nil {
				logrus.Warn("SaveMsg err :%s  ", err.Error())
				return
			}

			jsonBytes, err2 := json.Marshal(msg)
			if err2 != nil {
				logrus.Warn(" json.Marshal err :%s  ", err2.Error())
			}

			err2 = KafkaClient.ChatWriter.WriteMessages(ctx, kafka.Message{
				Key:   []byte(strconv.Itoa(config.Conf.Kafka.Partition)),
				Value: jsonBytes,
			})

			if err2 != nil {
				logrus.Warn(" KafkaClient.ChatWriter.WriteMessages err :%s  ", err2.Error())
				return
			}

		case proto.MsgTypePublic:
			// 群聊，投到对应房间的广播通道
			targetBucket := s.Bucket(transMsg.TargetId)
			if targetBucket == nil || targetBucket.rMap[transMsg.RoomId] == nil {
				logrus.Warn("target bucket or room does not exist")
				continue
			}

			parsedTime, err := time.Parse(time.RFC3339, transMsg.SendTimeStamp)
			if err != nil {
				logrus.Warn("ParseTime err :%s  ", err.Error())
			}

			// 用户在线
			msg := &proto.Message{
				MsgId:     tools.GetSnowflakeId(),
				Type:      transMsg.Type,
				Buffer:    []byte(transMsg.Buffer),
				MsgStatus: proto.MsgStatusSending,
				RoomId:    transMsg.RoomId,
				SenderId:  transMsg.SenderId,
				//TargetId:      targetSession.userId,
				SendTimeStamp: parsedTime,
				//ReceiveTimeStamp: time.Time{},
			}

			// redis持久化
			err = c.SaveMsg(msg)
			if err != nil {
				logrus.Warn("SaveMsg err :%s  ", err.Error())
				return
			}

			jsonBytes, err := json.Marshal(msg)
			if err != nil {
				logrus.Warn(" json.Marshal err :%s  ", err.Error())
			}

			err = KafkaClient.ChatWriter.WriteMessages(ctx, kafka.Message{
				Key:   []byte(strconv.Itoa(config.Conf.Kafka.Partition)),
				Value: jsonBytes,
			})
			if err != nil {
				logrus.Warn(" KafkaClient.ChatWriter.WriteMessages err :%s  ", err.Error())
				return
			}

		case proto.MsgTypeDisConnect:
			// 客户端断开连接
			logrus.Info("client disConnect")
			return
		default:
			logrus.Warnf("unknown msg type :%s", message)
			//return
		}
	}
}
