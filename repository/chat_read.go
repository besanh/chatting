package repository

import "github.com/besanh/chatting/model"

type (
	IChatRead interface {
		IPgRepo[model.ChatRead]
	}
	ChatRead struct {
		PgRepo[model.ChatRead]
	}
)

var ChatReadRepo IChatRead

func NewChatRead() IChatRead {
	return &ChatRead{}
}
