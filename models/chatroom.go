package models

import (
	"container/list"
	"github.com/google/uuid"
)

const archiveSize = 20 // 保存历史消息的条数
const chanSize = 10    // 通道buffer默认大小=10

const msgJoin = "[加入房间]"
const msgLeave = "[离开房间]"
const msgTyping = "[正在输入]"

// Room 聊天室
type Room struct {
	users          map[uid]chan Event
	userCount      int                    // 当前房间总人数
	publishChannel chan Event             // 聊天室推送channel
	archive        *list.List             // 历史记录 todo: 未持久化 重启失效
	archiveChan    chan chan []Event      // 通过接受chan来同步聊天内容
	joinChn        chan chan Subscription // 接收订阅事件的通道 用户加入聊天室后要把历史事件推送给用户
	leaveChn       chan uid               // 用户取消订阅通道 把通道中的历史事件释放并把用户从聊天室用户列表中删除
}

// NewRoom 初始化一个聊天室，不需要参数
func NewRoom() *Room {
	room := &Room{
		users:     map[uid]chan Event{}, // todo:为什么其他通道给分配内存，这个通道不分配内存？
		userCount: 0,

		publishChannel: make(chan Event, chanSize),
		archive:        list.New(),
		archiveChan:    make(chan chan []Event, chanSize),

		joinChn:  make(chan chan Subscription, chanSize),
		leaveChn: make(chan uid, chanSize),
	}

	go room.Serve()

	return room
}

// UserJoin 用户加入，公告给所有人
func (room *Room) UserJoin(user string) {
	room.publishChannel <- NewEvent(EventTypeJoin, user, msgJoin)
}

// UserLeave 用户离开
func (room *Room) UserLeave(user string) {
	room.publishChannel <- NewEvent(EventTypeLeave, user, msgLeave)
}

// UserSay 用户发送消息
func (room *Room) UserSay(user, msg string) {
	room.publishChannel <- NewEvent(EventTypeMsg, user, msg)
}

// Remove 用户从聊天室移除
func (room *Room) Remove(uid string) {
	room.leaveChn <- uid
}

// JoinRoom 用户订阅room
func (room *Room) JoinRoom(username string) Subscription {
	response := make(chan Subscription)
	room.joinChn <- response
	s := <-response
	s.Username = username
	return s
}

func (room *Room) GetArchive() []Event {
	ch := make(chan []Event)
	room.archiveChan <- ch
	return <-ch
}

// Serve 启动一个go程，监听这个room的通道
func (room *Room) Serve() {
	for {
		select {
		// 用户加入房间
		case ch := <-room.joinChn:
			room.userCount++
			chn := make(chan Event, chanSize) // 用户对应的接收消息的通道
			uid := uuid.New().String()
			room.users[uid] = chn // 这个收消息的通道和room的users绑定起来
			// 新增一个订阅，这个用户发消息的通道是上面这个chn；收消息通道是room的公告通道；离开时需要推送的通道也是room的用户离开通道
			ch <- Subscription{
				Id:                uid,
				Pipe:              chn,
				EmitChannel:       room.publishChannel,
				LeaveEventChannel: room.leaveChn,
			}
			// 然后告诉所有人我来了
			joinEvent := NewEvent(EventTypeJoin, uid, msgJoin)
			joinEvent.UserCount = room.userCount // todo:为什么要维护event里的userCount字段？起什么作用
			for _, userChannel := range room.users {
				userChannel <- joinEvent
			}

		// 取出	room.archiveChan 最前面的通道，然后把room的所以历史消息推送进去
		case arch := <-room.archiveChan:
			var events []Event // todo:为什么不能用new([]Event)
			for e := room.archive.Front(); e != nil; e = e.Next() {
				events = append(events, e.Value.(Event))
			}
			arch <- events

		// 公共channel里有消息
		case event := <-room.publishChannel:
			event.UserCount = room.userCount
			for _, userChannel := range room.users {
				userChannel <- event // 挨个推送消息
			}
			// 历史消息太多啦，删除最早的那条
			if room.archive.Len() > archiveSize {
				room.archive.Remove(room.archive.Front())
			}
			// 刚收到的event存进去
			room.archive.PushBack(event)

		// 有人想要退出房间
		case uid := <-room.leaveChn:
			if _, ok := room.users[uid]; ok { // 取user[uid]对应的channel时判断一下，别出错了不报错
				delete(room.users, uid) // 把这个uid和对应的channel从map里删去
				room.userCount--
			}
			// 公告：这个用户离开了
			leaveEvent := NewEvent(EventTypeLeave, uid, msgLeave)
			leaveEvent.UserCount = room.userCount
			for _, userChannel := range room.users {
				userChannel <- leaveEvent
			}
		}

	}
}
