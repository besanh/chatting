package model

import (
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

	ChatMember struct {
		*GBase
		bun.BaseModel `bun:"table:chat_member"`
		UserId        string    `json:"user_id" bun:"user_id,type:uuid,notnull"`
		JoinedAt      time.Time `json:"joined_at" bun:",nullzero,notnull,default:current_timestamp"`
		Role          string    `json:"role" bun:",notnull,default:'member'"`
	}

	ChatRead struct {
		*GBase
		bun.BaseModel `bun:"table:chat_reads,alias:cr"`
		UserId        string    `json:"user_id" bun:"user_id,pk,type:uuid,notnull"`
		LastRead      time.Time `json:"last_read" bun:",nullzero,notnull,default:current_timestamp"`
	}
)
