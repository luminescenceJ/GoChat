package connect

import (
	"Go-Chat/proto"
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

func (c *Connect) SaveMsg(msg *proto.Message) (err error) {
	key := "msg:" + strconv.FormatInt(msg.MsgId, 10)
	fields := map[string]interface{}{
		"msg_id":             msg.MsgId,
		"type":               msg.Type,
		"buffer":             msg.Buffer,
		"msg_status":         msg.MsgStatus,
		"room_id":            msg.RoomId,
		"sender_id":          msg.SenderId,
		"target_id":          msg.TargetId,
		"send_time_stamp":    msg.SendTimeStamp.Unix(),
		"receive_time_stamp": msg.ReceiveTimeStamp.Unix(),
	}
	// 写入 Hash
	if err = RedisClient.HMSet(key, fields).Err(); err != nil {
		return err
	}
	// 写入 sender/receiver 的 SortedSet
	score := float64(msg.SendTimeStamp.UnixNano())

	if err = RedisClient.ZAdd("user:"+strconv.Itoa(msg.SenderId)+":sent_msgs", redis.Z{
		Score:  score,
		Member: msg.MsgId,
	}).Err(); err != nil {
		return err
	}

	if msg.TargetId != 0 {
		if err = RedisClient.ZAdd("user:"+strconv.Itoa(msg.TargetId)+":recv_msgs", redis.Z{
			Score:  score,
			Member: msg.MsgId,
		}).Err(); err != nil {
			return err
		}
	}
	return nil

}

func (c *Connect) UpdateMsgStatus(msgId int64, status int) error {
	key := "msg:" + strconv.FormatInt(msgId, 10)
	return RedisClient.HSet(key, "msg_status", status).Err()
}

func (c *Connect) GetUserMessageIDs(userId int, mode string, offset, count int64) ([]int64, error) {
	key := "user:" + strconv.Itoa(userId) + ":" + mode // sent_msgs / recv_msgs

	ids, err := RedisClient.ZRange(key, offset, offset+count-1).Result()
	if err != nil {
		return nil, err
	}

	var res []int64
	for _, idStr := range ids {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			res = append(res, id)
		}
	}
	return res, nil
}

func (c *Connect) GetMessage(msgId int64) (*proto.Message, error) {
	key := "msg:" + strconv.FormatInt(msgId, 10)
	data, err := RedisClient.HGetAll(key).Result()
	if err != nil || len(data) == 0 {
		return nil, err
	}

	msg := &proto.Message{}

	if id, err := strconv.ParseInt(data["msg_id"], 10, 64); err == nil {
		msg.MsgId = id
	}
	msg.Type, _ = strconv.Atoi(data["type"])
	msg.MsgStatus, _ = strconv.Atoi(data["msg_status"])
	msg.RoomId, _ = strconv.Atoi(data["room_id"])
	msg.SenderId, _ = strconv.Atoi(data["sender_id"])
	msg.TargetId, _ = strconv.Atoi(data["target_id"])
	if ts, _ := strconv.ParseInt(data["send_time_stamp"], 10, 64); ts > 0 {
		msg.SendTimeStamp = time.Unix(ts, 0)
	}
	if ts, _ := strconv.ParseInt(data["receive_time_stamp"], 10, 64); ts > 0 {
		msg.ReceiveTimeStamp = time.Unix(ts, 0)
	}
	msg.Buffer = []byte(data["buffer"])

	return msg, nil
}
