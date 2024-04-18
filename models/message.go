package models

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

type Message struct {
	gorm.Model
	ID        uint      `json:"id"`
	UserId    int       `json:"user_id"`
	ToUserId  int       `json:"to_user_id"`
	RoomId    int       `json:"room_id"`
	Content   string    `json:"content"`
	ImageUrl  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SaveContent 将message存储到数据库里
func SaveContent(value interface{}) Message {
	var message Message

	message.UserId = value.(map[string]interface{})["user_id"].(int)
	message.ToUserId = value.(map[string]interface{})["to_user_id"].(int)
	message.Content = value.(map[string]interface{})["content"].(string)

	roomIdStr := value.(map[string]interface{})["room_id"].(string)
	message.RoomId, _ = strconv.Atoi(roomIdStr)

	if _, ok := value.(map[string]interface{})["image_url"]; ok {
		message.ImageUrl = value.(map[string]interface{})["image_url"].(string)
	}

	ChatDB.Create(&message)

	return message
}

// GetLimitMsg 获取指定房间中的100条历史消息
func GetLimitMsg(roomId string, offset int) []map[string]interface{} {
	var result []map[string]interface{}

	ChatDB.Model(&Message{}).
		Select("messages.*,users.username,users.avatar_id").
		Joins("INNER Join users on users.id=messages.user_id").
		Where("messages.room_id = ? and messages.to_user_id = ?", roomId, 0).
		Order("messages.id").
		Offset(offset).
		Limit(100).
		Scan(&result)

	return result
}

func GetLimitPrivateMsg(uid, toUId string, offset int) []map[string]interface{} {
	var result []map[string]interface{}

	ChatDB.Model(&Message{}).
		Select("messages.*,users.username,users.avatar_id").
		Joins("INNER Join users on users.id=messages.user_id").
		Where("messages.user_id = ? and messages.to_user_id = ?", uid, toUId).
		Where("messages.user_id = ? and messages.to_user_id = ?", toUId, uid).
		Order("messages.id").
		Offset(offset).
		Limit(100).
		Scan(&result)

	return result
}
