package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProfile(context *gin.Context) {
	if context.MustGet("account") == "Tomas" && context.MustGet("role") == "Member" {
		context.JSON(http.StatusOK, gin.H{
			"name":  "Tomas",
			"age":   23,
			"hobby": "music",
		})
		return
	}

	context.JSON(http.StatusNotFound, gin.H{
		"error": "can not find the record",
	})
}
