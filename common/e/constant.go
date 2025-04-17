package e

const (
	SuccessReplyCode = 0
	FailReplyCode    = 1

	SuccessReplyMsg    = "success"
	QueueName          = "gochat_queue"
	RedisBaseValidTime = 86400

	CurrentId   = "currentId"
	CurrentName = "currentName"

	CodeSuccess      = 0
	CodeFail         = 1
	CodeUnknownError = -1

	//RedisPrefix           = "gochat_"
	//RedisRoomPrefix       = "gochat_room_"
	//RedisRoomOnlinePrefix = "gochat_room_online_count_"
	//

	//MsgVersion      = 1
	//OpSingleSend    = 2 // single user
	//OpRoomSend      = 3 // send to room
	//OpRoomCountSend = 4 // get online user count
	//OpRoomInfoSend  = 5 // send info to room
	//OpBuildTcpConn  = 6 // build tcp conn
)

var MsgCodeMap = map[int]string{
	CodeUnknownError: "unKnow error",
	CodeSuccess:      "success",
	CodeFail:         "fail",
}
