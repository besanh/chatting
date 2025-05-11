package model

import (
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type (
	Chat struct {
		*GBase
		bun.BaseModel `bun:"table:chat,alias:chat"`
		Title         string        `json:"title" bun:"title,type:text,notnull"`
		IsGroup       bool          `json:"is_group" bun:"is_group,type:boolean,notnull,default:false"`
		ChatMember    []*ChatMember `json:"chat_member" bun:"rel:has-many,join:id=chat_id"`
	}

	PinelineCreateChatRequest struct {
		Title   string `json:"title" form:"required"`
		IsGroup bool   `json:"is_group" form:"required"`

		MemberIds []string `json:"member_ids" form:"required"`
	}

	ChatFilter struct {
		ChatId    string   `json:"chat_id"`
		MemberIds []string `json:"member_ids"`
	}
)

type (
	ChatMember struct {
		*GBase
		bun.BaseModel `bun:"table:chat_member,alias:cm"`
		ChatId        string    `json:"chat_id" bun:"chat_id,pk,type:uuid,notnull"`
		UserId        string    `json:"user_id" bun:"user_id,type:uuid,notnull"`
		JoinedAt      time.Time `json:"joined_at" bun:",nullzero,notnull,default:current_timestamp"`
		Role          string    `json:"role" bun:",notnull,default:'member'"`
		Chat          *Chat     `json:"chat" bun:"rel:belongs-to,join:chat_id=id"`
	}
)

type (
	ChatRead struct {
		*GBase
		bun.BaseModel `bun:"table:chat_reads,alias:cr"`
		UserId        string    `json:"user_id" bun:"user_id,pk,type:uuid,notnull"`
		LastRead      time.Time `json:"last_read" bun:",nullzero,notnull,default:current_timestamp"`
	}
)

func (m *PinelineCreateChatRequest) Validate() (err error) {
	if len(m.Title) < 1 {
		err = errors.New("title is empty")
	}

	if len(m.MemberIds) < 1 {
		err = errors.New("member_ids is empty")
	}

	return
}
