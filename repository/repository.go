package repository

import (
	"context"

	"github.com/besanh/chatbot_gpt/pkg/sqlclient"
)

var DBConn sqlclient.ISqlClientConn

func CreateTable(ctx context.Context, db sqlclient.ISqlClientConn, entity any) (err error) {
	_, err = db.GetDB().NewCreateTable().Model(entity).
		IfNotExists().
		Exec(ctx)
	return
}

func InitTables(ctx context.Context, dbConn sqlclient.ISqlClientConn) {
	// if err := CreateTable(ctx, dbConn, (*model.Example)(nil)); err != nil {
	// 	log.Error(err)
	// }
}
