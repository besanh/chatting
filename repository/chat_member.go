package repository

import "github.com/besanh/chatting/model"

type (
	IChatMember interface {
		IPgRepo[model.ChatMember]
	}
	ChatMember struct {
		PgRepo[model.ChatMember]
	}
)

var ChatMemberRepo IChatMember

func NewChatMember() IChatMember {
	return &ChatMember{}
}
