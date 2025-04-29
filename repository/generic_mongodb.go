package repository

import (
	"context"
	"errors"
	"time"

	"github.com/besanh/chatting/model"
	"github.com/besanh/chatting/pkg/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	IRepoGeneric[T model.GModel] interface {
		GetById(ctx context.Context, id string) (result *T, err error)
		Select(ctx context.Context, limit, offset int64, params ...Filter) (total int64, entities []*T, err error)
		CountFilter(ctx context.Context, filter Filter) (total int64, err error)
		GetCollection() *mongo.Collection

		// Insert
		Insert(ctx context.Context, entity *T) (err error)
		BulkInsert(ctx context.Context, entities []*T) (err error)

		// Update
		UpdateById(ctx context.Context, entity *T) (err error)
		BulkWriteUpdate(ctx context.Context, entities []*T) (err error)
		BulkUpdateByFilter(ctx context.Context, filter bson.M, entities []*T) (err error)
		BulkUpdateOneById(ctx context.Context, entities []*T) (err error)

		// Delete
		DeleteById(ctx context.Context, id string) (err error)
		BulkDeleteOneById(ctx context.Context, entities []*T) (err error)
		BulkWriteDelete(ctx context.Context, entities []*T) (err error)
		BulkDeleteManyByFilter(ctx context.Context, filter bson.M) (err error)

		// Transaction
		StartSession() (mongo.Session, error)
		EndSession(ctx context.Context, session mongo.Session)
		StartTransaction(session mongo.Session) error
		AbortTransaction(ctx context.Context, session mongo.Session) error
		CommitTransaction(ctx context.Context, session mongo.Session) error
	}

	RepoGeneric[T model.GModel] struct {
		DB         mongodb.IMongoDBClient
		Collection string
	}

	Filter struct {
		Key   string
		Value any
	}
)

func NewGenericRepo[T model.GModel](db mongodb.IMongoDBClient, collection string) IRepoGeneric[T] {
	return &RepoGeneric[T]{
		DB:         db,
		Collection: collection,
	}
}

func (repo *RepoGeneric[T]) GetById(ctx context.Context, id string) (result *T, err error) {
	result = new(T)
	filters := make(bson.D, 0)
	filters = append(filters, bson.E{Key: "id", Value: id})
	err = repo.DB.Collection(repo.Collection).FindOne(ctx, filters).Decode(result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	return
}

func (repo *RepoGeneric[T]) Insert(ctx context.Context, entity *T) (err error) {
	result, errTmp := repo.DB.Collection(repo.Collection).InsertOne(ctx, entity)
	if result == nil || result.InsertedID == nil {
		err = errors.New("insert mongo failed")
	} else if errTmp != nil {
		err = errTmp
	}

	return
}

func (repo *RepoGeneric[T]) BulkInsert(ctx context.Context, entities []*T) (err error) {
	docs := make([]any, 0)
	for _, entity := range entities {
		(*entity).SetCreatedAt(time.Now())
		(*entity).SetUpdatedAt(time.Now())
		docs = append(docs, entity)
	}

	_, err = repo.DB.Collection(repo.Collection).InsertMany(ctx, docs, &options.InsertManyOptions{})
	return
}

func (repo *RepoGeneric[T]) UpdateById(ctx context.Context, entity *T) (err error) {
	(*entity).SetUpdatedAt(time.Now())
	result, err := repo.DB.Collection(repo.Collection).UpdateByID(ctx, (*entity).GetId(), entity, &options.UpdateOptions{})
	if result.MatchedCount == 0 {
		err = errors.New("update failed")
	}
	return
}

/*
 * @use bulk write when each entity has its own filter
 */
func (repo *RepoGeneric[T]) BulkWriteUpdate(ctx context.Context, entities []*T) (err error) {
	operations := make([]mongo.WriteModel, 0, len(entities))
	for _, entity := range entities {
		(*entity).SetUpdatedAt(time.Now())
		filter := bson.M{"id": (*entity).GetId()}

		updateEntity := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(bson.M{"$set": entity})

		operations = append(operations, updateEntity)
	}

	result, err := repo.DB.Collection(repo.Collection).BulkWrite(ctx, operations, &options.BulkWriteOptions{})
	if result.MatchedCount == 0 {
		err = errors.New("bulk write update failed")
	}

	return
}

/*
 * @use case: bulk update when each entity has the same filter and the slice is small
 */
func (repo *RepoGeneric[T]) BulkUpdateByFilter(ctx context.Context, filter bson.M, entities []*T) (err error) {
	for _, entity := range entities {
		(*entity).SetUpdatedAt(time.Now())
	}
	_, err = repo.DB.Collection(repo.Collection).UpdateMany(ctx, filter, bson.M{"$set": entities}, &options.UpdateOptions{})
	return
}

/*
 * @use case: update a small slice that follows id
 */
func (repo *RepoGeneric[T]) BulkUpdateOneById(ctx context.Context, entities []*T) (err error) {
	for _, entity := range entities {
		(*entity).SetUpdatedAt(time.Now())
		_, err = repo.DB.Collection(repo.Collection).UpdateOne(ctx, bson.M{"id": (*entity).GetId()}, bson.M{"$set": entities}, &options.UpdateOptions{})
	}

	return
}

func (repo *RepoGeneric[T]) DeleteById(ctx context.Context, id string) (err error) {
	result, err := repo.DB.Collection(repo.Collection).DeleteOne(ctx, bson.M{"id": id})
	if result.DeletedCount == 0 {
		err = errors.New("delete failed")
	}
	return
}

/*
 * @use case: delete a massive slice
 */
func (repo *RepoGeneric[T]) BulkWriteDelete(ctx context.Context, entities []*T) (err error) {
	operations := make([]mongo.WriteModel, 0, len(entities))
	for _, entity := range entities {
		filter := bson.M{"id": (*entity).GetId()}

		deleteEntity := mongo.NewDeleteOneModel().
			SetFilter(filter)

		operations = append(operations, deleteEntity)
	}

	result, err := repo.DB.Collection(repo.Collection).BulkWrite(ctx, operations, &options.BulkWriteOptions{})
	if result.MatchedCount == 0 {
		err = errors.New("bulk write delete failed")
	}

	return
}

/*
 * @use case: delete a small or medium slice as filter
 */
func (repo *RepoGeneric[T]) BulkDeleteManyByFilter(ctx context.Context, filter bson.M) (err error) {
	result, err := repo.DB.Collection(repo.Collection).DeleteMany(ctx, filter)
	if result.DeletedCount == 0 {
		err = errors.New("delete failed")
	}
	return
}

/*
 * @use case: delete a small or medium slice
 */
func (repo *RepoGeneric[T]) BulkDeleteOneById(ctx context.Context, entities []*T) (err error) {
	for _, entity := range entities {
		_, err = repo.DB.Collection(repo.Collection).DeleteMany(ctx, bson.M{"id": (*entity).GetId()})
	}
	return
}

func (repo *RepoGeneric[T]) Select(ctx context.Context, limit, offset int64, params ...Filter) (total int64, entities []*T, err error) {
	entities = make([]*T, 0)
	filters := make(bson.D, 0)
	for _, param := range params {
		filters = append(filters, primitive.E{Key: param.Key, Value: param.Value})
	}
	var cur *mongo.Cursor
	cur, err = repo.DB.Collection(repo.Collection).Find(ctx, filters, options.Find().SetLimit(limit).SetSkip(offset))
	if err == mongo.ErrNoDocuments {
		return 0, nil, nil
	} else if err != nil {
		return
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		entity := new(T)
		if err := cur.Decode(entity); err != nil {
			return 0, nil, err
		}
		entities = append(entities, entity)
	}
	if err := cur.Err(); err != nil {
		return 0, nil, err
	}
	total, err = repo.DB.Collection(repo.Collection).CountDocuments(ctx, filters)
	if err != nil {
		return 0, nil, err
	}
	return
}

func (repo *RepoGeneric[T]) CountFilter(ctx context.Context, filter Filter) (total int64, err error) {
	total, err = repo.DB.Collection(repo.Collection).CountDocuments(ctx, filter)
	return
}

func (repo *RepoGeneric[T]) GetCollection() *mongo.Collection {
	return repo.DB.Collection(repo.Collection)
}

func (repo *RepoGeneric[T]) StartSession() (mongo.Session, error) {
	return repo.DB.DB().Client().StartSession()
}

func (repo *RepoGeneric[T]) EndSession(ctx context.Context, session mongo.Session) {
	session.EndSession(ctx)
}

func (repo *RepoGeneric[T]) StartTransaction(session mongo.Session) error {
	return session.StartTransaction()
}

func (repo *RepoGeneric[T]) AbortTransaction(ctx context.Context, session mongo.Session) error {
	return session.AbortTransaction(ctx)
}

func (repo *RepoGeneric[T]) CommitTransaction(ctx context.Context, session mongo.Session) error {
	return session.CommitTransaction(ctx)
}
