package repository

import (
	"context"
	"database/sql"

	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/pkg/sqlclient"
	"github.com/uptrace/bun"
)

type (
	IChat interface {
		IPgRepo[model.Chat]
		GetChatExist(ctx context.Context, db sqlclient.ISqlClientConn, chatFilter model.ChatFilter) (result *model.Chat, err error)
	}
	Chat struct {
		PgRepo[model.Chat]
	}
)

var ChatRepo IChat

func NewChat() IChat {
	return &Chat{}
}

func (repo *Chat) GetChatExist(ctx context.Context, db sqlclient.ISqlClientConn, chatFilter model.ChatFilter) (result *model.Chat, err error) {
	result = &model.Chat{}
	query := db.GetDB().NewSelect().
		Model(result).
		Relation("ChatMember")
	if len(chatFilter.ChatId) > 0 {
		query.Where("id = ?", chatFilter.ChatId)
	}
	if len(chatFilter.MemberIds) > 0 {
		query.Where("cm.user_id IN (?)", bun.In(chatFilter.MemberIds))
	}

	err = query.Scan(ctx)
	if err != nil && err != sql.ErrNoRows {
		return
	}
	return
}
