package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var wsChan = make(chan WsPayload)
var clients = make(map[*websocket.Conn]string)

// WsJsonResponse defines the response sent back from websocket
type WsJsonResponse struct {
	Action            string   `json:"action"`
	Message           string   `json:"message"`
	MessageType       string   `json:"message_type"`
	MessageFrom       string   `json:"message_from"`
	MessageFromAvatar string   `json:"message_from_avatar"`
	MessageTime       string   `json:"message_time"`
	ConnectedUsers    []string `json:"connected_users"`
}

// WsPayload defines the websocket request from the client
type WsPayload struct {
	Action     string          `json:"action"`
	Username   string          `json:"username"`
	UserAvatar string          `json:"userAvatar"`
	Message    string          `json:"message"`
	Conn       *websocket.Conn `json:"-"`
}

func main() {
	g := gin.Default()
	g.Use(gin.Recovery())
	g.GET("/socket/:name", SocketHandler)
	g.Run(":8082")
}
func SocketHandler(c *gin.Context) {
	userName := c.Param("name")
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	//var isExists bool
	//
	//for _, s := range clients {
	//	if userName == s {
	//		isExists = true
	//		break
	//	}
	//}

	//if !isExists {
	clients[ws] = userName

	go ListenToWsChan()

	go func() {
		var playLoad WsPayload
		defer func() {
			_ = playLoad.Conn.Close()
		}()

		for true {
			err1 := ws.ReadJSON(&playLoad)
			if err1 != nil {
				break
			} else {
				playLoad.Conn = ws
				wsChan <- playLoad
			}
		}
	}()
	//}

}

func ListenToWsChan() {
	var response WsJsonResponse

	for true {
		user := <-wsChan

		switch user.Action {
		case "Init":
			clients[user.Conn] = user.Username
			users := GetOnlineUserList()
			response.ConnectedUsers = users
			response.Action = "Init"
			response.Message = user.Username + "已連線！"
			response.MessageType = "Init"
			response.MessageFromAvatar = user.UserAvatar
			fmt.Println(user.Username + "已連線！")
			BroadCastToAll(response)
		case "SendMsg":
			response.Action = "SendMsg"
			response.Message = user.Username + "：" + user.Message
			response.MessageTime = time.Now().Format("2006/01/02 15:04:05")
			response.MessageType = "SendMsg"
			response.MessageFrom = user.Username
			response.MessageFromAvatar = user.UserAvatar
			fmt.Println(user.Username + "説：" + user.Message)
			BroadCastToAll(response)
		case "Left":
			delete(clients, user.Conn)
			response.Action = "Left"
			response.Message = user.Username + "已離開！"
			response.MessageType = "Left"
			fmt.Println(user.Username + "已離開！")
			response.ConnectedUsers = GetOnlineUserList()
			BroadCastToAll(response)
		}
	}
}

func GetOnlineUserList() []string {
	userList := make([]string, 0)
	for _, x := range clients {
		userList = append(userList, x)
	}
	return userList
}

func BroadCastToAll(response WsJsonResponse) {
	for conn, _ := range clients {
		err := conn.WriteJSON(response)
		if err != nil {
			conn.Close()
			delete(clients, conn)
		}
	}
}
