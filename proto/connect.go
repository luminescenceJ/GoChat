package proto

// 用于前后端消息传输
type TransMessage struct {
	Type int `json:"type"`
	//SendTimeStamp time.Time `json:"sendTimeStamp"`
	Buffer []byte `json:"buffer"`
}

type EnterRoomRequest struct {
	UserId   int    `json:"user_id"`
	RoomId   int    `json:"room_id"`
	ServerId string `json:"serverId"`
}

type EnterRoomResponse struct {
	Code int `json:"code"`
}
