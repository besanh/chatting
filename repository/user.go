package repository

import (
	"github.com/besanh/chatting/model"
)

type (
	IUser interface {
		IPgRepo[model.User]
	}
	User struct {
		PgRepo[model.User]
	}
)

var UserRepo IUser

func NewUser() IUser {
	return &User{}
}
