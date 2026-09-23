package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

// UserRepository 用户数据访问层。
type UserRepository struct {
	coll *mongo.Collection
}

// NewUserRepository 构造注入。
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{coll: db.Collection("users")}
}

func (r *UserRepository) Create(ctx context.Context, u *model.User) error {
	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, u); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&u); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": u.ID}, u)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateCounts(ctx context.Context, id primitive.ObjectID, incFollow, incFollower int64) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"follow_count": incFollow, "follower_count": incFollower},
	})
	if err != nil {
		return fmt.Errorf("update user counts: %w", err)
	}
	return nil
}

func (r *UserRepository) ListByIDs(ctx context.Context, ids []primitive.ObjectID) ([]model.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("list users by ids: %w", err)
	}
	defer cursor.Close(ctx)
	users := []model.User{}
	if err := cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return users, nil
}

// CreateFollow 关注（唯一索引防重）。
func (r *UserRepository) CreateFollow(ctx context.Context, f *model.Follow) error {
	if f.ID.IsZero() {
		f.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.Database().Collection("follows").InsertOne(ctx, f); err != nil {
		return fmt.Errorf("insert follow: %w", err)
	}
	return nil
}

func (r *UserRepository) DeleteFollow(ctx context.Context, followerID, followeeID primitive.ObjectID) error {
	res, err := r.coll.Database().Collection("follows").DeleteOne(ctx, bson.M{"follower_id": followerID, "followee_id": followeeID})
	if err != nil {
		return fmt.Errorf("delete follow: %w", err)
	}
	if res.DeletedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *UserRepository) IsFollowing(ctx context.Context, followerID, followeeID primitive.ObjectID) (bool, error) {
	n, err := r.coll.Database().Collection("follows").CountDocuments(ctx, bson.M{"follower_id": followerID, "followee_id": followeeID})
	if err != nil {
		return false, fmt.Errorf("count follow: %w", err)
	}
	return n > 0, nil
}

func (r *UserRepository) ListFollowingIDs(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error) {
	filter := bson.M{"follower_id": userID}
	total, err := r.coll.Database().Collection("follows").CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count following: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Database().Collection("follows").Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list following: %w", err)
	}
	defer cursor.Close(ctx)
	var follows []model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, 0, fmt.Errorf("decode follows: %w", err)
	}
	ids := make([]primitive.ObjectID, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FolloweeID)
	}
	return ids, total, nil
}

func (r *UserRepository) ListFollowerIDs(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error) {
	filter := bson.M{"followee_id": userID}
	total, err := r.coll.Database().Collection("follows").CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count followers: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Database().Collection("follows").Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list followers: %w", err)
	}
	defer cursor.Close(ctx)
	var follows []model.Follow
	if err := cursor.All(ctx, &follows); err != nil {
		return nil, 0, fmt.Errorf("decode follows: %w", err)
	}
	ids := make([]primitive.ObjectID, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.FollowerID)
	}
	return ids, total, nil
}
