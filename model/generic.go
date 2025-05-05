package model

import (
	"time"

	"github.com/google/uuid"
)

type (
	GModel interface {
		GetId() string
		SetId(id string)
		SetCreatedAt(t time.Time)
		SetUpdatedAt(t time.Time)
	}

	GBase struct {
		Id        string    `json:"id" bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
		CreatedAt time.Time `bun:"created_at,notnull"`
		UpdatedAt time.Time `bun:"updated_at,notnull"`
	}
)

func (b *GBase) GetId() string {
	return b.Id
}

func (b *GBase) SetId(id string) {
	b.Id = id
}

func (b *GBase) SetCreatedAt(t time.Time) {
	b.CreatedAt = t
}

func (b *GBase) SetUpdatedAt(t time.Time) {
	b.UpdatedAt = t
}

func InitPgBase() *GBase {
	return &GBase{
		Id:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
