package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/pkg/sqlclient"
	"github.com/uptrace/bun"
)

type IPgRepo[T model.GModel] interface {
	GetById(ctx context.Context, db sqlclient.ISqlClientConn, id string) (*T, error)
	Insert(ctx context.Context, db sqlclient.ISqlClientConn, entity T) error
	Update(ctx context.Context, db sqlclient.ISqlClientConn, entity T) error
	Delete(ctx context.Context, db sqlclient.ISqlClientConn, id string) error
	CreateTable(ctx context.Context, db sqlclient.ISqlClientConn) (err error)
	SelectByQuery(ctx context.Context, db sqlclient.ISqlClientConn, params []model.Param, limit int, offset int) (entries *[]T, total int, err error)
	BulkInsert(ctx context.Context, db sqlclient.ISqlClientConn, entities []T) error
	TxSelectByQuery(ctx context.Context, tx bun.Tx, params []model.Param, limit, offset int) (entities *[]T, total int, err error)
	TxInsert(ctx context.Context, tx bun.Tx, entity T) (err error)
}

type PgRepo[T model.GModel] struct {
}

func NewRepo[T model.GModel]() IPgRepo[T] {
	return &PgRepo[T]{}
}

func (r *PgRepo[T]) CreateTable(ctx context.Context, db sqlclient.ISqlClientConn) (err error) {
	query := db.GetDB().NewCreateTable().Model((*T)(nil)).
		IfNotExists()
	_, err = query.
		Exec(ctx)
	return
}

func (r *PgRepo[T]) GetById(ctx context.Context, db sqlclient.ISqlClientConn, id string) (entity *T, err error) {
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

func (r *PgRepo[T]) Insert(ctx context.Context, db sqlclient.ISqlClientConn, entity T) (err error) {
	entity.SetCreatedAt(time.Now())
	_, err = db.GetDB().NewInsert().
		Model(&entity).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) Update(ctx context.Context, db sqlclient.ISqlClientConn, entity T) (err error) {
	entity.SetUpdatedAt(time.Now())
	_, err = db.GetDB().NewUpdate().
		Model(&entity).
		Where("id = ?", entity.GetId()).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) Delete(ctx context.Context, db sqlclient.ISqlClientConn, id string) (err error) {
	_, err = db.GetDB().NewDelete().
		Model((*T)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) SelectByQuery(ctx context.Context, db sqlclient.ISqlClientConn, params []model.Param, limit int, offset int) (entries *[]T, total int, err error) {
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

func (r *PgRepo[T]) BulkInsert(ctx context.Context, db sqlclient.ISqlClientConn, entities []T) (err error) {
	_, err = db.GetDB().NewInsert().
		Model(&entities).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) TxSelectByQuery(ctx context.Context, tx bun.Tx, params []model.Param, limit, offset int) (entries *[]T, total int, err error) {
	entries = new([]T)
	query := tx.NewSelect().
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

func (r *PgRepo[T]) TxInsert(ctx context.Context, tx bun.Tx, entity T) (err error) {
	entity.SetCreatedAt(time.Now())
	_, err = tx.NewInsert().
		Model(&entity).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) TxUpdate(ctx context.Context, tx bun.Tx, entity T) (err error) {
	entity.SetUpdatedAt(time.Now())
	_, err = tx.NewUpdate().
		Model(&entity).
		Where("id = ?", entity.GetId()).
		Exec(ctx)
	return
}

func (r *PgRepo[T]) TxDelete(ctx context.Context, tx bun.Tx, id string) (err error) {
	_, err = tx.NewDelete().
		Model((*T)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return
}
