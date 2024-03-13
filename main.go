package main

import (
	"bytes"
	"log"
	"net/http"

	"github.com/horsekitlin/go-gin-postgresql-example/conf"
	"github.com/horsekitlin/go-gin-postgresql-example/models"
	"github.com/horsekitlin/go-gin-postgresql-example/routers"
	"github.com/spf13/viper"
)

func init() {

	viper.SetConfigType("json") // 設置配置文件的類型

	if err := viper.ReadConfig(bytes.NewBuffer(conf.AppJsonConfig)); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("no such config file")
		} else {
			// Config file was found but another error was produced
			log.Println("read config error")
		}
		log.Fatal(err) // 讀取配置文件失敗致命錯誤
	}

	models.InitDB()
}

func main() {
	routersInit := routers.InitRouter()
	port := viper.GetString(`app.port`)

	log.Println("監聽端口", "http://127.0.0.1:"+port)

	http.ListenAndServe(":"+port, routersInit)
}
