package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
)

type (
	Chat struct {
		*GBase
		bun.BaseModel `bun:"table:chat"`
		Title         string `json:"title" bun:"title,type:text,notnull"`
		IsGroup       bool   `json:"is_group" bun:"is_group,type:boolean,notnull,default:false"`
	}

	PinelineCreateChatRequest struct {
		Title   string       `json:"title" form:"required"`
		IsGroup sql.NullBool `json:"is_group" form:"required"`

		MemberIds []string `json:"member_ids" form:"required"`
	}
)

type (
	ChatMember struct {
		*GBase
		bun.BaseModel `bun:"table:chat_member"`
		ChatId        string    `json:"chat_id" bun:"chat_id,pk,type:uuid,notnull"`
		UserId        string    `json:"user_id" bun:"user_id,type:uuid,notnull"`
		JoinedAt      time.Time `json:"joined_at" bun:",nullzero,notnull,default:current_timestamp"`
		Role          string    `json:"role" bun:",notnull,default:'member'"`
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
