package repository

import (
	"context"
	"fmt"

	"github.com/besanh/chatting/model"
)

type (
	IUser interface {
		IPgRepo[model.User]
		ClearRefreshToken(ctx context.Context, id string) (err error)
	}
	User struct {
		PgRepo[model.User]
	}
)

var UserRepo IUser

func NewUser() IUser {
	return &User{}
}

func (repo *User) ClearRefreshToken(ctx context.Context, id string) (err error) {
	query := DBConn.GetDB().
		NewUpdate().
		Model((*model.User)(nil)).
		Set("refresh_token_encrypted", "").
		Where("id = ?", id)

	result, err := query.Exec(ctx)
	if total, _ := result.RowsAffected(); total == 0 {
		err = fmt.Errorf("clear refresh token failed: %w", err)
	}

	return
}
