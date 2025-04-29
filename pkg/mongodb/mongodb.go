package mongodb

import (
	"context"
	"fmt"
	"time"

	log "github.com/besanh/logger/logging/slog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type (
	IMongoDBClient interface {
		Connect() (err error)
		Disconnect()
		DB() *mongo.Database
		Collection(mollection string) *mongo.Collection
		SetCollectionNames(arr []string)
		GetCollectionNames() []string
		Ping()
	}
	MongoDBClient struct {
		Config           MongoDBConfig
		ConnectionString string
		Credential       options.Credential
		Database         string
		CollectionNames  []string
		Client           *mongo.Client
	}

	MongoDBConfig struct {
		Username      string
		Password      string
		Host          string
		Port          int
		Database      string
		Ssl           bool
		DefaultAuthDb string
	}
)

func NewMongoDBClient(config MongoDBConfig) (IMongoDBClient, error) {
	connectionString := fmt.Sprintf("mongodb://%s:%d/%s?authSource=%s&readPreference=primary&ssl=%v&directConnection=true", config.Host, config.Port, config.Database, config.DefaultAuthDb, config.Ssl)
	credential := options.Credential{
		Username: config.Username,
		Password: config.Password,
	}

	db := &MongoDBClient{
		Credential:       credential,
		Database:         config.Database,
		ConnectionString: connectionString,
	}

	if err := db.Connect(); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return db, nil
}

func (m *MongoDBClient) Connect() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(m.ConnectionString).SetAuth(m.Credential))
	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	m.Client = client

	return
}

func (m *MongoDBClient) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	m.Client.Disconnect(ctx)
}

func (m *MongoDBClient) DB() *mongo.Database {
	return m.Client.Database(m.Database)
}

func (m *MongoDBClient) Collection(mollection string) *mongo.Collection {
	return m.Client.Database(m.Database).Collection(mollection)
}

func (m *MongoDBClient) SetCollectionNames(arr []string) {
	m.CollectionNames = arr
}
func (m *MongoDBClient) GetCollectionNames() []string {
	return m.CollectionNames
}

func (m *MongoDBClient) Ping() {
	if err := m.Client.Ping(context.Background(), nil); err != nil {
		log.Error("error: %v", err)
	}
	log.Debug("Connected to MongoDB")
}
