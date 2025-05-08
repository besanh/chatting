package repository

import "github.com/besanh/chatting/model"

type (
	IChat interface {
		IPgRepo[model.Chat]
	}
	Chat struct {
		PgRepo[model.Chat]
	}
)

var ChatRepo IChat

func NewChat() IChat {
	return &Chat{}
}
