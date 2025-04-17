package proto

type LoginRequest struct {
	Name     string
	Password string
}

type LoginResponse struct {
	Code      int
	AuthToken string
}

type RegisterRequest struct {
	Name     string
	Password string
}

type RegisterReply struct {
	Code      int
	AuthToken string
}

type LogoutRequest struct {
	AuthToken string
}

type LogoutResponse struct {
	Code int
}

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
