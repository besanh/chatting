package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	GMongoModel interface {
		GetId() string
		SetId(id string)
		SetCreatedAt(t time.Time)
		SetUpdatedAt(t time.Time)
	}

	GMongoBase struct {
		Id        primitive.ObjectID `json:"id" bson:"_id"`
		CreatedAt time.Time          `json:"created_at" bson:"created_at"`
		UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	}
)

func (b *GMongoBase) GetId() string {
	return b.Id.Hex()
}

func (b *GMongoBase) SetId(id string) {
	b.Id, _ = primitive.ObjectIDFromHex(id)
}

func (b *GMongoBase) SetCreatedAt(t time.Time) {
	b.CreatedAt = t
}

func (b *GMongoBase) SetUpdatedAt(t time.Time) {
	b.UpdatedAt = t
}

func InitBase() *GMongoBase {
	return &GMongoBase{
		Id:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
