package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/pkg/sqlclient"
)

type IRepo[T model.GModel] interface {
	GetById(ctx context.Context, db sqlclient.ISqlClientConn, id string) (*T, error)
	Insert(ctx context.Context, db sqlclient.ISqlClientConn, entity T) error
	Update(ctx context.Context, db sqlclient.ISqlClientConn, entity T) error
	Delete(ctx context.Context, db sqlclient.ISqlClientConn, id string) error
	CreateTable(ctx context.Context, db sqlclient.ISqlClientConn) (err error)
	SelectByQuery(ctx context.Context, db sqlclient.ISqlClientConn, params []model.Param, limit int, offset int) (entries *[]T, total int, err error)
	BulkInsert(ctx context.Context, db sqlclient.ISqlClientConn, entities []T) error
}

type Repo[T model.GModel] struct {
}

func NewRepo[T model.GModel]() IRepo[T] {
	return &Repo[T]{}
}

func (r *Repo[T]) CreateTable(ctx context.Context, db sqlclient.ISqlClientConn) (err error) {
	query := db.GetDB().NewCreateTable().Model((*T)(nil)).
		IfNotExists()
	_, err = query.
		Exec(ctx)
	return
}

func (r *Repo[T]) GetById(ctx context.Context, db sqlclient.ISqlClientConn, id string) (entity *T, err error) {
	entity = new(T)
	err = db.GetDB().NewSelect().
		Model(entity).
		Where("id = ?", id).
		Limit(1).
		Scan(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *Repo[T]) Insert(ctx context.Context, db sqlclient.ISqlClientConn, entity T) (err error) {
	entity.SetCreatedAt(time.Now())
	_, err = db.GetDB().NewInsert().
		Model(&entity).
		Exec(ctx)
	return
}

func (r *Repo[T]) Update(ctx context.Context, db sqlclient.ISqlClientConn, entity T) (err error) {
	entity.SetUpdatedAt(time.Now())
	_, err = db.GetDB().NewUpdate().
		Model(&entity).
		Where("id = ?", entity.GetId()).
		Exec(ctx)
	return
}

func (r *Repo[T]) Delete(ctx context.Context, db sqlclient.ISqlClientConn, id string) (err error) {
	_, err = db.GetDB().NewDelete().
		Model((*T)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return
}

func (r *Repo[T]) SelectByQuery(ctx context.Context, db sqlclient.ISqlClientConn, params []model.Param, limit int, offset int) (entries *[]T, total int, err error) {
	entries = new([]T)
	query := db.GetDB().NewSelect().
		Model(entries).
		Limit(limit).
		Offset(offset)
	for _, param := range params {
		qb := param.BuildQuery()
		if len(qb) > 0 {
			query.Where(qb, param.Value)
		}
	}
	total, err = query.ScanAndCount(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, 0, err
	}
	return entries, total, nil
}

func (r *Repo[T]) BulkInsert(ctx context.Context, db sqlclient.ISqlClientConn, entities []T) (err error) {
	_, err = db.GetDB().NewInsert().
		Model(&entities).
		Exec(ctx)
	return
}
