package connect

import (
	"Go-Chat/proto"
	"sync"
	"sync/atomic"
)

/*
Bucket 是用于记录Room和Session的结构，通过分片设计来避免用户增改与群发消息冲突的锁竞争
routines 可以通过多协程提高cpu利用率，因为不同群发消息之间是隔离的，只需要上读锁，使用原子加routinesNum做到无锁状态下的轮询实现负载均衡
*/
type Bucket struct {
	Lock        sync.RWMutex
	sMap        map[int]*Session      // map useId to user Session
	rMap        map[int]*Room         // map userId to specific Room
	routines    []chan *proto.Message // goroutine for public message sending
	routinesNum uint64                // concurrent polling counter
	options     BucketOptions
}

type BucketOptions struct {
	SessionSize   int    // 用户个数
	RoomSize      int    // 聊天室
	RoutineAmount uint64 // 处理群发消息的并发协程数
	RoutineSize   int    // 每个协程所能积压的最大消息
}

func NewBucket(op BucketOptions) *Bucket {
	bucket := new(Bucket)
	bucket.sMap = make(map[int]*Session, op.SessionSize)
	bucket.rMap = make(map[int]*Room, op.RoomSize)
	bucket.routines = make([]chan *proto.Message, op.RoutineAmount)
	bucket.options = op
	bucket.routinesNum = 0
	for i := uint64(0); i < op.RoutineAmount; i++ {
		c := make(chan *proto.Message, op.RoutineSize)
		bucket.routines[i] = c
		go bucket.Listen(c) // 使用多协程监听广播消息,推送到具体房间
	}
	return bucket
}

// 根据房间号获得实例
func (b *Bucket) GetRoom(rid int) (room *Room) {
	b.Lock.RLock()
	room, _ = b.rMap[rid]
	b.Lock.RUnlock()
	return
}

// 根据用户id获得实例
func (b *Bucket) GetSession(userId int) (ss *Session) {
	b.Lock.RLock()
	ss = b.sMap[userId]
	b.Lock.RUnlock()
	return
}

// 监听广播通道并发送
func (b *Bucket) Listen(ch chan *proto.Message) {
	for {
		var (
			msg  *proto.Message
			room *Room
		)
		msg = <-ch
		if room = b.GetRoom(msg.RoomId); room != nil {
			room.GroupMessaging(msg)
		}
	}
}

// 新增用户Session
func (b *Bucket) Put(userId int, roomId int, ss *Session) (err error) {
	var (
		room *Room
		ok   bool
	)
	b.Lock.Lock()

	if roomId != NoRoom {
		if room, ok = b.rMap[roomId]; !ok {
			room = NewRoom(roomId)
			b.rMap[roomId] = room
		}
		ss.Room = room
	}
	ss.userId = userId
	b.sMap[userId] = ss

	b.Lock.Unlock()
	if room != nil {
		err = room.Put(ss)
	}
	return
}

// 删除用户Session
func (b *Bucket) Delete(ss *Session) {
	var (
		ok   bool
		room *Room
	)
	b.Lock.RLock()
	if ss, ok = b.sMap[ss.userId]; ok {
		room = b.sMap[ss.userId].Room
		//delete from bucket
		delete(b.sMap, ss.userId)
	}
	if room != nil && room.Remove(ss) {
		// if room empty delete,will mark room.drop is true
		if room.Status == true {
			delete(b.rMap, room.Id)
		}
	}
	b.Lock.RUnlock()
}

// 根据原子加实现负载均衡减少锁竞争
func (b *Bucket) BroadcastRoom(msg *proto.Message) {
	num := atomic.AddUint64(&b.routinesNum, 1) % b.options.RoutineAmount
	b.routines[num] <- msg
}
