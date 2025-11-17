package model

import (
	"context"
	"oauth2/config"
)

type User struct {
	ID       uint   `gorm:"primary_key" json:"id"`
	Username string `gorm:"size:255" json:"username"`
	Password string `gorm:"size:255" json:"password"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

func (u *User) TableName() string {
	return "user"
}

func (u *User) Authentication(ctx context.Context, username, password string) (userID uint, err error) {
	if config.GetCfg().AuthMode == "db" {
		u := new(User)
		if err := db.WithContext(ctx).Where("username = ? AND password = ?", username, password).First(u).Error; err != nil {
			return 0, err
		}
		if u != nil {
			userID = u.ID
		}
	}
	if config.GetCfg().AuthMode == "ldap" {

	}
	return
}
