package routers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var tokens []string

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func InitRouter() *gin.Engine {
	route := gin.New()
	route.Use(gin.Logger())
	route.Use(gin.Recovery())

	route.POST("/login", gin.BasicAuth(gin.Accounts{
		"admin": "secret",
	}), func(context *gin.Context) {
		token, _ := randomHex(20)
		tokens = append(tokens, token)

		context.JSON(http.StatusOK, gin.H{
			"token": token,
		})
	})

	route.GET("/", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "hello world!",
		})
	})

	route.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	route.GET("/home", func(context *gin.Context) {
		bearerToken := context.Request.Header.Get("Authorization")
		reqToken := strings.Split(bearerToken, " ")[1]
		for _, token := range tokens {
			if token == reqToken {
				context.JSON(http.StatusOK, gin.H{
					"data": "resource data",
				})
				return
			}
		}
		context.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
	})

	route.GET("/resource", gin.BasicAuth(gin.Accounts{
		"admin": "secret",
	}), func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"data": "resource data",
		})
	})

	return route
}
