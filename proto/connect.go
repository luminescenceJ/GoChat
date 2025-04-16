package proto

// 用于前后端消息传输
type TransMessage struct {
	Type int `json:"type"`
	//SendTimeStamp time.Time `json:"sendTimeStamp"`
	Buffer []byte `json:"buffer"`
}
