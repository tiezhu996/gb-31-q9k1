package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/petsocial/petsocial/internal/constants"
)

// DB 数据库句柄聚合。
type DB struct {
	Client *mongo.Client
	Mongo  *mongo.Database
}

// Connect 连接 MongoDB 并初始化索引。
func Connect(ctx context.Context, uri, dbName string, logger *slog.Logger) (*DB, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}
	db := client.Database(dbName)
	if err := ensureIndexes(ctx, db); err != nil {
		return nil, fmt.Errorf("ensure indexes: %w", err)
	}
	logger.Info(fmt.Sprintf(constants.LogDBConnected, dbName))
	return &DB{Client: client, Mongo: db}, nil
}

// Close 关闭连接。
func (d *DB) Close(ctx context.Context) error {
	return d.Client.Disconnect(ctx)
}

func ensureIndexes(ctx context.Context, db *mongo.Database) error {
	collections := map[string][]mongo.IndexModel{
		"users": {
			{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true)},
			{Keys: bson.D{{Key: "city", Value: 1}}},
		},
		"follows": {
			{Keys: bson.D{{Key: "follower_id", Value: 1}, {Key: "followee_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"pets": {
			{Keys: bson.D{{Key: "owner_id", Value: 1}}},
			{Keys: bson.D{{Key: "species", Value: 1}}},
			{Keys: bson.D{{Key: "city", Value: 1}}},
		},
		"pet_follows": {
			{Keys: bson.D{{Key: "pet_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"posts": {
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "author_id", Value: 1}, {Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "topics", Value: 1}}},
			{Keys: bson.D{{Key: "city", Value: 1}}},
			{Keys: bson.D{{Key: "status", Value: 1}}},
		},
		"comments": {
			{Keys: bson.D{{Key: "post_id", Value: 1}, {Key: "created_at", Value: 1}}},
		},
		"interactions": {
			{Keys: bson.D{{Key: "post_id", Value: 1}, {Key: "user_id", Value: 1}, {Key: "type", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"topics": {
			{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		},
		"meetups": {
			{Keys: bson.D{{Key: "city", Value: 1}, {Key: "status", Value: 1}}},
			{Keys: bson.D{{Key: "creator_id", Value: 1}}},
			{Keys: bson.D{{Key: "participants.user_id", Value: 1}}},
		},
		"chat_messages": {
			{Keys: bson.D{{Key: "from_id", Value: 1}, {Key: "to_id", Value: 1}, {Key: "created_at", Value: 1}}},
			{Keys: bson.D{{Key: "to_id", Value: 1}, {Key: "read", Value: 1}}},
		},
		"audit_logs": {
			{Keys: bson.D{{Key: "created_at", Value: -1}}},
			{Keys: bson.D{{Key: "username", Value: 1}}},
			{Keys: bson.D{{Key: "action", Value: 1}}},
		},
	}
	for name, models := range collections {
		if _, err := db.Collection(name).Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("create indexes for %s: %w", name, err)
		}
	}
	return nil
}
