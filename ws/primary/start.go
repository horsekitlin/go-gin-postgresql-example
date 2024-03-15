package primary

import (
	"github.com/horsekitlin/go-gin-postgresql-example/ws"

	"github.com/gin-gonic/gin"
)

// 定義 serve 的映射關係
var serveMap = map[string]ws.ServeInterface{
	"Serve": &ws.Serve{},
}

func Create() ws.ServeInterface {
	return serveMap["Serve"]
}

func Start(gin *gin.Context) {
	Create().RunWs(gin)
}
