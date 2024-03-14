package routers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/horsekitlin/go-gin-postgresql-example/routers/api"
	v1 "github.com/horsekitlin/go-gin-postgresql-example/routers/api/v1"
	"github.com/horsekitlin/go-gin-postgresql-example/ws/primary"
)

var (
	tokens []string
	i      interface{}
)

// var tokens []string

// jwt secret key
var jwtSecret = []byte("secret")

// custom claims
type Claims struct {
	Account string `json:"account"`
	Role    string `json:"role"`
	jwt.StandardClaims
}

// validate JWT
func AuthRequired(context *gin.Context) {
	auth := context.GetHeader("Authorization")
	token := strings.Split(auth, "Bearer ")[1]
	fmt.Println("token: " + token)
	// parse and validate token for six things:
	// validationErrorMalformed => token is malformed
	// validationErrorUnverifiable => token could not be verified because of signing problems
	// validationErrorSignatureInvalid => signature validation failed
	// validationErrorExpired => exp validation failed
	// validationErrorNotValidYet => nbf validation failed
	// validationErrorIssuedAt => iat validation failed
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecret, nil
	})

	if err != nil {
		var message string
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				message = "token is malformed"
			} else if ve.Errors&jwt.ValidationErrorUnverifiable != 0 {
				message = "token could not be verified because of signing problems"
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				message = "signature validation failed"
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				message = "token is expired"
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				message = "token is not yet valid before sometime"
			} else {
				message = "can not handle this token"
			}
		}
		context.JSON(http.StatusUnauthorized, gin.H{
			"error": message,
		})
		context.Abort()
		return
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		fmt.Println("account:", claims.Account)
		fmt.Println("role:", claims.Role)
		context.Set("account", claims.Account)
		context.Set("role", claims.Role)
		context.Next()
	} else {
		context.Abort()
		return
	}
}

func SocketHandler(c *gin.Context) {
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()

	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Message Type: %d, Message: %s\n", msgType, string(msg))
		err = ws.WriteJSON(struct {
			Reply string `json:"reply"`
		}{
			Reply: "Echo...",
		})
		if err != nil {
			panic(err)
		}
	}
}

func InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/ping", api.GetPing)
	router.GET("/home", api.GetHome)
	router.GET("/resource", gin.BasicAuth(gin.Accounts{
		"admin": "secret",
	}), api.GetResource)
	router.POST("/login", api.GetAuth)

	router.GET("/ws", primary.Start)

	apiv1 := router.Group("/api/v1")
	apiv1.Use(AuthRequired)
	{
		apiv1.GET("/member/profile", v1.GetProfile)
	}

	return router
}
