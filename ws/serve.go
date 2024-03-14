package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Serve struct {
	ServeInterface
}

func (serve *Serve) RunWs(gin *gin.Context) {
	Run(gin)
}

func (serve *Serve) GetOnlineUserCount() int {
	return GetOnlineUserCount()
}

func (serve *Serve) GetOnlineRoomUserCount(roomId int) int {
	return GetOnlineRoomUserCount(roomId)
}

// 客户端連接詳情
type wsClients struct {
	Conn *websocket.Conn `json:"conn"`

	RemoteAddr string `json:"remote_addr"`

	Uid float64 `json:"uid"`

	Username string `json:"username"`

	RoomId string `json:"room_id"`

	AvatarId string `json:"avatar_id"`

	ToUser interface{} `json:"to_user"`
}

// client & serve 的消息體
type msg struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

// 變量定義初始化
var (
	wsUpgrader = websocket.Upgrader{}

	clientMsg = msg{}

	mutex = sync.Mutex{}

	rooms = [roomCount + 1][]wsClients{}

	privateChat = []wsClients{}
)

// 定義消息類型
const msgTypeOnline = 1        // 上線
const msgTypeOffline = 2       // 離線
const msgTypeSend = 3          // 消息發送
const msgTypeGetOnlineUser = 4 // 獲取用户列表
const msgTypePrivateChat = 5   // 私聊

const roomCount = 6 // 房間總數

func Run(gin *gin.Context) {

	// @see https://github.com/gorilla/websocket/issues/523
	wsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }

	c, _ := wsUpgrader.Upgrade(gin.Writer, gin.Request, nil)

	defer c.Close()

	mainProcess(c)
}

// 主程序，負責循環讀取客户端消息跟消息的發送
func mainProcess(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		serveMsgStr := message

		// 處理心跳響應 , heartbeat為與客户端約定的值
		if string(serveMsgStr) == `heartbeat` {
			c.WriteMessage(websocket.TextMessage, []byte(`{"status":0,"data":"heartbeat ok"}`))
			continue
		}
		fmt.Printf("message: " + string(message))

		json.Unmarshal(message, &clientMsg)
		// log.Println("來自客户端的消息", clientMsg,c.RemoteAddr())
		if clientMsg.Data == nil {
			return
			//mainProcess(c)
		}

		if err != nil { // 離線通知
			log.Println("ReadMessage error1", err)
			disconnect(c)
			c.Close()
			return
		}

		if clientMsg.Status == msgTypeOnline { // 進入房間，建立連接
			handleConnClients(c)
			serveMsgStr = formatServeMsgStr(msgTypeOnline)
		}

		if clientMsg.Status == msgTypePrivateChat {
			// 處理私聊
			serveMsgStr = formatServeMsgStr(msgTypePrivateChat)
			toC := findToUserCoonClient()
			if toC != nil {
				toC.(wsClients).Conn.WriteMessage(websocket.TextMessage, serveMsgStr)
			}
		}

		if clientMsg.Status == msgTypeSend { // 消息發送
			serveMsgStr = formatServeMsgStr(msgTypeSend)
		}

		if clientMsg.Status == msgTypeGetOnlineUser {
			serveMsgStr = formatServeMsgStr(msgTypeGetOnlineUser)
			c.WriteMessage(websocket.TextMessage, serveMsgStr)
			continue
		}

		//log.Println("serveMsgStr", string(serveMsgStr))
		if clientMsg.Status == msgTypeSend || clientMsg.Status == msgTypeOnline {
			notify(c, string(serveMsgStr))
		}
	}
}

// 獲取私聊的用户連接
func findToUserCoonClient() interface{} {
	_, roomIdInt := getRoomId()

	toUserUid := clientMsg.Data.(map[string]interface{})["to_uid"].(string)

	for _, c := range rooms[roomIdInt] {
		stringUid := strconv.FormatFloat(c.Uid, 'f', -1, 64)
		if stringUid == toUserUid {
			return c
		}
	}

	return nil
}

// 處理建立連接的用户
func handleConnClients(c *websocket.Conn) {
	roomId, roomIdInt := getRoomId()

	for cKey, wcl := range rooms[roomIdInt] {
		if wcl.Uid == clientMsg.Data.(map[string]interface{})["uid"].(float64) {
			mutex.Lock()
			// 通知當前用户下線
			wcl.Conn.WriteMessage(websocket.TextMessage, []byte(`{"status":-1,"data":[]}`))
			rooms[roomIdInt] = append(rooms[roomIdInt][:cKey], rooms[roomIdInt][cKey+1:]...)
			wcl.Conn.Close()
			mutex.Unlock()
		}
	}

	mutex.Lock()
	rooms[roomIdInt] = append(rooms[roomIdInt], wsClients{
		Conn:       c,
		RemoteAddr: c.RemoteAddr().String(),
		Uid:        clientMsg.Data.(map[string]interface{})["uid"].(float64),
		Username:   clientMsg.Data.(map[string]interface{})["username"].(string),
		RoomId:     roomId,
		AvatarId:   clientMsg.Data.(map[string]interface{})["avatar_id"].(string),
	})
	mutex.Unlock()
}

// 統一消息發放
func notify(conn *websocket.Conn, msg string) {
	_, roomIdInt := getRoomId()
	for _, con := range rooms[roomIdInt] {
		if con.RemoteAddr != conn.RemoteAddr().String() {
			con.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

// 離線通知
func disconnect(conn *websocket.Conn) {
	_, roomIdInt := getRoomId()
	for index, con := range rooms[roomIdInt] {
		if con.RemoteAddr == conn.RemoteAddr().String() {
			data := map[string]interface{}{
				"username": con.Username,
				"uid":      con.Uid,
				"time":     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
			}

			jsonStrServeMsg := msg{
				Status: msgTypeOffline,
				Data:   data,
			}
			serveMsgStr, _ := json.Marshal(jsonStrServeMsg)

			disMsg := string(serveMsgStr)

			mutex.Lock()
			rooms[roomIdInt] = append(rooms[roomIdInt][:index], rooms[roomIdInt][index+1:]...)
			con.Conn.Close()
			mutex.Unlock()
			notify(conn, disMsg)
		}
	}
}

// 格式化傳送給客户端的消息數據
func formatServeMsgStr(status int) []byte {

	roomId, roomIdInt := getRoomId()

	data := map[string]interface{}{
		"username": clientMsg.Data.(map[string]interface{})["username"].(string),
		"uid":      clientMsg.Data.(map[string]interface{})["uid"].(float64),
		"room_id":  roomId,
		"time":     time.Now().UnixNano() / 1e6, // 13位  10位 => now.Unix()
	}

	if status == msgTypeSend || status == msgTypePrivateChat {
		data["avatar_id"] = clientMsg.Data.(map[string]interface{})["avatar_id"].(string)
		data["content"] = clientMsg.Data.(map[string]interface{})["content"].(string)

		// toUidStr := clientMsg.Data.(map[string]interface{})["to_uid"].(string)
		// toUid, _ := strconv.Atoi(toUidStr)

		// 保存消息
		// stringUid := strconv.FormatFloat(data["uid"].(float64), 'f', -1, 64)
		// intUid, _ := strconv.Atoi(stringUid)

		// if _, ok := clientMsg.Data.(map[string]interface{})["image_url"]; ok {
		// 	// 存在圖片
		// 	models.SaveContent(map[string]interface{}{
		// 		"user_id":    intUid,
		// 		"to_user_id": toUid,
		// 		"content":    data["content"],
		// 		"room_id":    data["room_id"],
		// 		"image_url":  clientMsg.Data.(map[string]interface{})["image_url"].(string),
		// 	})
		// } else {
		// 	models.SaveContent(map[string]interface{}{
		// 		"user_id":    intUid,
		// 		"to_user_id": toUid,
		// 		"room_id":    data["room_id"],
		// 		"content":    data["content"],
		// 	})
		// }

	}

	if status == msgTypeGetOnlineUser {
		data["count"] = GetOnlineRoomUserCount(roomIdInt)
		data["list"] = onLineUserList(roomIdInt)
	}

	jsonStrServeMsg := msg{
		Status: status,
		Data:   data,
	}
	serveMsgStr, _ := json.Marshal(jsonStrServeMsg)

	return serveMsgStr
}

func getRoomId() (string, int) {
	roomId := clientMsg.Data.(map[string]interface{})["room_id"].(string)

	roomIdInt, _ := strconv.Atoi(roomId)
	return roomId, roomIdInt
}

// 獲取在線用户列表
func onLineUserList(roomId int) []wsClients {
	return rooms[roomId]
}

// =======================對外方法=====================================

func GetOnlineUserCount() int {
	num := 0
	for i := 1; i <= roomCount; i++ {
		num = num + GetOnlineRoomUserCount(i)
	}
	return num
}

func GetOnlineRoomUserCount(roomId int) int {
	return len(rooms[roomId])
}
