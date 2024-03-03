package main

import (
	"fmt"

	"github.com/horsekitlin/go-gin-postgresql-example/routers"
)

func main() {
	routersInit := routers.InitRouter()
	endPoint := fmt.Sprintf(":%d", 8080)
	routersInit.Run(endPoint)
}
