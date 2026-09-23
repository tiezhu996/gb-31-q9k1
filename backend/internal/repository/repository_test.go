package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// testDB 连接测试库；未配置 MONGO_TEST_URI 时跳过。
func testDB(t *testing.T) *mongo.Database {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set, skip repository test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect test mongo: %v", err)
	}
	db := client.Database("petsocial_test_" + time.Now().Format("150405"))
	t.Cleanup(func() {
		_ = db.Drop(context.Background())
		_ = client.Disconnect(context.Background())
	})
	return db
}

func ensureTestIndexes(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.Collection("follows").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "follower_id", Value: 1}, {Key: "followee_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	_, _ = db.Collection("interactions").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "post_id", Value: 1}, {Key: "user_id", Value: 1}, {Key: "type", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

func cleanCollection(t *testing.T, db *mongo.Database, name string) {
	t.Helper()
	_, _ = db.Collection(name).DeleteMany(context.Background(), bson.M{})
}
