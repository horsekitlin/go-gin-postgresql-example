package models

import (
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ChatDB *gorm.DB

func InitDB() *gorm.DB {
	dsn := viper.GetString(`database.dsn`)
	ChatDB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return ChatDB
}
