package proto

type ConnectRequest struct {
	AuthToken string `json:"authToken"`
	RoomId    int    `json:"roomId"`
	ServerId  string `json:"serverId"`
}

type DisConnectRequest struct {
	RoomId int
	UserId int
}

type ConnectReply struct {
	UserId int
}

type DisConnectReply struct {
	Has bool
}
