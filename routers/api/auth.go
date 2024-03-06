package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var tokens []string

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

func GetAuth(context *gin.Context) {
	// validate request body
	var body struct {
		Account  string
		Password string
	}
	err := context.ShouldBindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// check account and password is correct
	if body.Account == "Tomas" && body.Password == "123456" {
		now := time.Now()
		jwtId := body.Account + strconv.FormatInt(now.Unix(), 10)
		role := "Member"

		// set claims and sign
		claims := Claims{
			Account: body.Account,
			Role:    role,
			StandardClaims: jwt.StandardClaims{
				Audience:  body.Account,
				ExpiresAt: now.Add(2000 * time.Second).Unix(),
				Id:        jwtId,
				IssuedAt:  now.Unix(),
				Issuer:    "ginJWT",
				NotBefore: now.Add(1 * time.Second).Unix(),
				Subject:   body.Account,
			},
		}
		tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token, err := tokenClaims.SignedString(jwtSecret)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"token": token,
		})
		return
	}

	// incorrect account or password
	context.JSON(http.StatusUnauthorized, gin.H{
		"message": "Unauthorized",
	})
}
