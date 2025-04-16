package connect

import (
	"Go-Chat/common/e"
	"Go-Chat/proto"
	"github.com/sirupsen/logrus"
	"sync"
)

const NoRoom = -1

type Room struct {
	Id          int // 房间Id
	OnlineCount int // 在线人数
	rLock       sync.RWMutex
	Status      bool
	next        *Session
}

func NewRoom(roomId int) *Room {
	room := new(Room)
	room.Id = roomId
	room.Status = true
	room.next = nil
	room.OnlineCount = 0
	return room
}

// 将用户加入Room
func (r *Room) Put(s *Session) (err error) {
	r.rLock.Lock()
	defer r.rLock.Unlock()
	if r.Status == false {
		return e.Error_ROOM_DROP
	}
	// 将新的session插入头部
	if r.next != nil {
		r.next.Prev = s
	}
	s.Next = r.next
	s.Prev = nil
	r.next = s
	r.OnlineCount++
	return
}

// 从Room中删除Session，返回Room的状态
func (r *Room) Remove(s *Session) bool {
	r.rLock.RLock()
	defer r.rLock.RUnlock()
	if s.Next != nil {
		s.Next.Prev = s.Prev
	}
	if s.Prev != nil {
		s.Prev.Next = s.Next
	} else {
		// 头部的session
		r.next = s.Next
	}
	r.OnlineCount--
	if r.OnlineCount <= 0 {
		r.Status = false
	}
	return r.Status
}

// 群发消息
func (r *Room) GroupMessaging(msg *proto.Message) {
	r.rLock.RLock()
	defer r.rLock.RUnlock()
	for ch := r.next; ch != nil; ch = ch.Next {
		if err := ch.SendMsg(msg); err != nil {
			logrus.Infof("GroupMessaging err:%s", err.Error())
		}
	}
}
