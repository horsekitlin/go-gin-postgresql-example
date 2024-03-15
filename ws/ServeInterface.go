package ws

import "github.com/gin-gonic/gin"

type ServeInterface interface {
	RunWs(gin *gin.Context)
}
