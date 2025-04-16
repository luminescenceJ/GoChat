package proto

import "time"

// todo : 多媒体聊天室:视频、语音

// MsgStatus 消息状态机流转
const (
	MsgStatusSending = iota // 客户端显示"发送中"
	MsgStatusRetry          // 客户端重试发送
	MsgStatusSent           // 服务端确认送达
	MsgStatusRead           // 接收方已读
)

// MsgType 消息类型:私聊、群聊
const (
	MsgTypePublic = iota
	MsgTypePrivate
	MsgTypeConnect
	MsgTypeDisConnect
)

type TransMsg struct {
	//MsgId            int64  `json:"msg_id"`
	Type   int    `json:"type"`
	Buffer string `json:"buffer"` //  []byte 无法直接反序列化字符串
	//MsgStatus        int    `json:"msg_status,omitempty"`
	RoomId        int    `json:"room_id,omitempty"`
	SenderId      int    `json:"sender_id"` // 前端必须传此字段
	TargetId      int    `json:"target_id,omitempty"`
	SendTimeStamp string `json:"send_time_stamp,omitempty"` // 时间格式统一
	//ReceiveTimeStamp string `json:"receive_time_stamp,omitempty"`
}

// 最基本的消息体结构
type Message struct {
	// 元数据
	MsgId            int64     `json:"msg_id"`                       // 消息自增全局ID 雪花算法生成
	Type             int       `json:"type"`                         // 消息类型：私聊/群聊
	Buffer           []byte    `json:"buffer"`                       // 实际的消息载荷
	MsgStatus        int       `json:"msg_status,omitempty"`         // 消息的状态
	RoomId           int       `json:"room_id,omitempty"`            // 聊天室Id
	SenderId         int       `json:"sender_id"`                    // 发送方
	TargetId         int       `json:"target_id,omitempty"`          // 接收方
	SendTimeStamp    time.Time `json:"send_time_stamp,omitempty"`    // 发送时间
	ReceiveTimeStamp time.Time `json:"receive_time_stamp,omitempty"` // 接收时间
}

//// GroupMessage 用于群发消息的多协程监听通道
//type GroupMessage struct {
//	RoomId int
//	Msg    Message
//}

//// TcpMsg 用于tcp下的流协议处理
//type TcpMsg struct {
//	Version string `json:"version"` // 协议版本
//	//Operation int    `json:"operation"`
//	Body []byte `json:"body"` // 二进制数据
//}
