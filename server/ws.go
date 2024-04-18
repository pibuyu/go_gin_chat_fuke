package server

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var Websocket = &WS{
	upgrader: &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	},
}

type WS struct {
	upgrader *websocket.Upgrader
}

func (ws *WS) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			panic(err)
		}

		// 加入房间
		archiveEvents := Room.GetArchive() // 获取历史消息
		Room.UserJoin(name)                // 通知大家说这个用户来了
		roomSubscripetion := Room.JoinRoom(name)
		defer roomSubscripetion.Leave()

		for _, archiveEvent := range archiveEvents {
			err := conn.WriteJSON(archiveEvent)
			if err != nil {
				return //用户断开连接
			}
		}

		// 监听用户事件，发送给聊天室
		newMessages := make(chan string)
		go func() {
			var res = struct {
				Msg string `json:"msg"`
			}{}
			for {
				err2 := conn.ReadJSON(&res)
				if err2 != nil { // 断开连接
					close(newMessages)
					return
				}
				// 没问题就把消息发出去
				newMessages <- res.Msg
			}
		}()

		for {
			select {
			// 接收消息
			case event := <-roomSubscripetion.Pipe:
				err := conn.WriteJSON(event)
				if err != nil {
					return
				}

			//发消息
			case msg, ok := <-newMessages:
				if !ok {
					return
				}
				roomSubscripetion.Say(msg)
			}
		}

	}
}
