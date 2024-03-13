package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Id       uint `gorm:"primaryKey"`
	Username string
	Email    string `gorm:"unique"`
}

func createUser(value interface{}) User {
	var u User
	u.Username = value.(map[string]interface{})["username"].(string)
	u.Email = value.(map[string]interface{})["email"].(string)
	ChatDB.Create(&u)
	return u
}
