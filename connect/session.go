package connect

import (
	"Go-Chat/proto"
	"github.com/gorilla/websocket"
)

/*
为什么要设计cache？
如果在多个goroutine同时调用发送消息，直接写WebSocket连接可能会导致并发问题，因为WebSocket的连接写入不是线程安全的。
通过一个单独的channel，可以让所有发送请求都通过这个channel，由writePump单独处理，确保同一时间只有一个写入操作，避免并发冲突。

WebSocket 的吞吐量主要受限于网络 I/O，通过broadcast的缓存加入，使得消息推送是非阻塞的
即使客户端暂时没有读取，消息仍然可以存储在 chan 里，减少 WebSocket 阻塞的可能性
此外，可以通过size限制用户最多积累消息，超过后可能会丢弃旧消息，避免OOM
*/
type Session struct {
	Room    *Room
	Next    *Session
	Prev    *Session
	cache   chan *proto.Message // 广播缓存通道
	userId  int
	conn_ws *websocket.Conn // 实际通信的websocket
}

func NewSession(size int) *Session {
	s := new(Session)
	s.cache = make(chan *proto.Message, size) // 最多可以保留size条记录
	// todo : 加入拉去历史信息
	// todo : mongo 存储信息
	return s
}

func (s *Session) SendMsg(msg *proto.Message) (err error) {
	select {
	case s.cache <- msg:
	default:
		// 加入redis，防止消息挤压导致丢失
	}
	return
}
