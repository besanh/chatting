package repository

import (
	"context"

	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/pkg/sqlclient"
)

var DBConn sqlclient.ISqlClientConn

func CreateTable(ctx context.Context, db sqlclient.ISqlClientConn, entity any) (err error) {
	_, err = db.GetDB().NewCreateTable().Model(entity).
		IfNotExists().
		Exec(ctx)
	return
}

func InitTables(ctx context.Context, dbConn sqlclient.ISqlClientConn) {
	if err := CreateTable(ctx, dbConn, (*model.User)(nil)); err != nil {
		panic(err)
	}
	if err := CreateTable(ctx, dbConn, (*model.Chat)(nil)); err != nil {
		panic(err)
	}
	if err := CreateTable(ctx, dbConn, (*model.ChatMember)(nil)); err != nil {
		panic(err)
	}
	if err := CreateTable(ctx, dbConn, (*model.ChatRead)(nil)); err != nil {
		panic(err)
	}
}

func InitRepositories() {
	UserRepo = NewUser()
	ChatRepo = NewChat()
	ChatMemberRepo = NewChatMember()
	ChatReadRepo = NewChatRead()
}

func InitColumn(ctx context.Context, db sqlclient.ISqlClientConn) {
}
