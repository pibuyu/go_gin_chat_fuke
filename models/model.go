package models

import "time"

type uid = string

const (
	EventTypeMsg    = "event-msg"    // 用户发言
	EventTypeSystem = "event-system" // 系统信息推送 如房间人数
	EventTypeJoin   = "event-join"   // 用户加入
	EventTypeTyping = "event-typing" // 用户正在输入
	EventTypeLeave  = "event-leave"  // 用户离开
	EventTypeImage  = "event-image"  // todo 消息图片
)

type Event struct {
	Type      string `json:"type"`      // 事件类型
	User      string `json:"user"`      // 用户名
	Timestamp int64  `json:"timestamp"` // 时间戳
	Text      string `json:"text"`      // 事件内容,实际上就是msg
	UserCount int    `json:"userCount"` // 房间用户数
}

func NewEvent(typ string, user, msg string) Event {
	return Event{
		Type:      typ,
		User:      user,
		Timestamp: time.Now().UnixNano() / 1e6,
		Text:      msg,
	}
}

func (event *Event) Create() bool {
	err = ChatDB.Create(&event).Error
	//err = ChatDB.Save(&event).Error
	if err != nil {
		return false
	}
	return true
}

// Subscription 用户订阅
type Subscription struct {
	Id                string
	Username          string
	Pipe              <-chan Event // 只接受通道，接受消息的通道
	EmitChannel       chan Event   // 发送消息的通道
	LeaveEventChannel chan uid     // 用户离开时间的通道
}

// Leave 用户离开了
func (subscription *Subscription) Leave() {
	subscription.LeaveEventChannel <- subscription.Id
}

// Say 发送消息
func (subscription *Subscription) Say(msg string) {
	subscription.EmitChannel <- NewEvent(EventTypeMsg, subscription.Username, msg)
}
