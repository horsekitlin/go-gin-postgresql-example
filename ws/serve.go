package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Serve struct {
	ServeInterface
}

func (serve *Serve) RunWs(gin *gin.Context) {
	Run(gin)
}

// 變量定義初始化
var (
	wsUpgrader = websocket.Upgrader{}
)

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
		messageType, message, err := c.ReadMessage()
		serveMsgStr := message

		if err != nil {
			c.WriteMessage(websocket.TextMessage, []byte(`{"status":0,"data":"get error"}`))
			continue
		}
		// 處理心跳響應 , heartbeat為與客户端約定的值
		if string(serveMsgStr) == `heartbeat` {
			c.WriteMessage(websocket.TextMessage, []byte(`{"status":0,"data":"heartbeat ok"}`))
			continue
		} else {
			c.WriteMessage(messageType, message)
			continue
		}
	}
}
