package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

// PostRepository 动态数据访问层。
type PostRepository struct {
	coll *mongo.Collection
}

// NewPostRepository 构造注入。
func NewPostRepository(db *mongo.Database) *PostRepository {
	return &PostRepository{coll: db.Collection("posts")}
}

func (r *PostRepository) Create(ctx context.Context, p *model.Post) error {
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, p); err != nil {
		return fmt.Errorf("insert post: %w", err)
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Post, error) {
	var p model.Post
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find post by id: %w", err)
	}
	return &p, nil
}

func (r *PostRepository) Update(ctx context.Context, p *model.Post) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PostRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status constants.PostStatus, reason string) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"status": status, "reject_reason": reason, "updated_at": now()},
	})
	if err != nil {
		return fmt.Errorf("update post status: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

// List 通用动态列表（复用：feed 发现页/同城页/用户主页/话题页均调用本方法）。
func (r *PostRepository) List(ctx context.Context, filter bson.M, page, pageSize int) ([]model.Post, int64, error) {
	filter = cloneFilter(filter)
	if _, ok := filter["status"]; !ok {
		filter["status"] = constants.PostStatusApproved
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}
	defer cursor.Close(ctx)
	posts := []model.Post{}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, 0, fmt.Errorf("decode posts: %w", err)
	}
	return posts, total, nil
}

func (r *PostRepository) ListByAuthor(ctx context.Context, authorID primitive.ObjectID, page, pageSize int) ([]model.Post, int64, error) {
	return r.List(ctx, bson.M{"author_id": authorID}, page, pageSize)
}

func (r *PostRepository) ListByTopic(ctx context.Context, topic string, page, pageSize int) ([]model.Post, int64, error) {
	return r.List(ctx, bson.M{"topics": topic}, page, pageSize)
}

func (r *PostRepository) ListByCity(ctx context.Context, city string, page, pageSize int) ([]model.Post, int64, error) {
	return r.List(ctx, bson.M{"city": city}, page, pageSize)
}

func (r *PostRepository) ListByIDs(ctx context.Context, ids []primitive.ObjectID, page, pageSize int) ([]model.Post, int64, error) {
	if len(ids) == 0 {
		return []model.Post{}, 0, nil
	}
	filter := bson.M{"_id": bson.M{"$in": ids}, "status": constants.PostStatusApproved}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count posts by ids: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list posts by ids: %w", err)
	}
	defer cursor.Close(ctx)
	posts := []model.Post{}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, 0, fmt.Errorf("decode posts by ids: %w", err)
	}
	return posts, total, nil
}

func (r *PostRepository) UpdateCounts(ctx context.Context, id primitive.ObjectID, incLike, incComment, incFavorite, incForward int64) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"like_count": incLike, "comment_count": incComment, "favorite_count": incFavorite, "forward_count": incForward},
	})
	if err != nil {
		return fmt.Errorf("update post counts: %w", err)
	}
	return nil
}

func (r *PostRepository) IncTopicCount(ctx context.Context, topic string, inc int64) error {
	_, err := r.coll.Database().Collection("topics").UpdateOne(ctx, bson.M{"name": topic}, bson.M{"$inc": bson.M{"post_count": inc}})
	if err != nil {
		return fmt.Errorf("inc topic count: %w", err)
	}
	return nil
}

// Comment 相关。
func (r *PostRepository) CreateComment(ctx context.Context, c *model.Comment) error {
	if c.ID.IsZero() {
		c.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.Database().Collection("comments").InsertOne(ctx, c); err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}
	return nil
}

func (r *PostRepository) ListComments(ctx context.Context, postID primitive.ObjectID, page, pageSize int) ([]model.Comment, int64, error) {
	filter := bson.M{"post_id": postID}
	coll := r.coll.Database().Collection("comments")
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count comments: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": 1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list comments: %w", err)
	}
	defer cursor.Close(ctx)
	comments := []model.Comment{}
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, fmt.Errorf("decode comments: %w", err)
	}
	return comments, total, nil
}

func (r *PostRepository) DeleteComment(ctx context.Context, commentID, postID, authorID primitive.ObjectID) error {
	res, err := r.coll.Database().Collection("comments").DeleteOne(ctx, bson.M{"_id": commentID, "post_id": postID, "author_id": authorID})
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if res.DeletedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PostRepository) FindComment(ctx context.Context, commentID primitive.ObjectID) (*model.Comment, error) {
	var c model.Comment
	if err := r.coll.Database().Collection("comments").FindOne(ctx, bson.M{"_id": commentID}).Decode(&c); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find comment: %w", err)
	}
	return &c, nil
}

// Interaction 相关。
func (r *PostRepository) CreateInteraction(ctx context.Context, in *model.Interaction) error {
	if in.ID.IsZero() {
		in.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.Database().Collection("interactions").InsertOne(ctx, in); err != nil {
		return fmt.Errorf("insert interaction: %w", err)
	}
	return nil
}

func (r *PostRepository) DeleteInteraction(ctx context.Context, postID, userID primitive.ObjectID, typ constants.InteractionType) error {
	res, err := r.coll.Database().Collection("interactions").DeleteOne(ctx, bson.M{"post_id": postID, "user_id": userID, "type": typ})
	if err != nil {
		return fmt.Errorf("delete interaction: %w", err)
	}
	if res.DeletedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PostRepository) HasInteraction(ctx context.Context, postID, userID primitive.ObjectID, typ constants.InteractionType) (bool, error) {
	n, err := r.coll.Database().Collection("interactions").CountDocuments(ctx, bson.M{"post_id": postID, "user_id": userID, "type": typ})
	if err != nil {
		return false, fmt.Errorf("count interaction: %w", err)
	}
	return n > 0, nil
}

func (r *PostRepository) ListInteractions(ctx context.Context, postID primitive.ObjectID, typ constants.InteractionType, page, pageSize int) ([]model.Interaction, int64, error) {
	filter := bson.M{"post_id": postID}
	if typ != "" {
		filter["type"] = typ
	}
	coll := r.coll.Database().Collection("interactions")
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count interactions: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list interactions: %w", err)
	}
	defer cursor.Close(ctx)
	items := []model.Interaction{}
	if err := cursor.All(ctx, &items); err != nil {
		return nil, 0, fmt.Errorf("decode interactions: %w", err)
	}
	return items, total, nil
}

// Topic 相关。
func (r *PostRepository) ListTopics(ctx context.Context, page, pageSize int) ([]model.Topic, int64, error) {
	coll := r.coll.Database().Collection("topics")
	total, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("count topics: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"post_count": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list topics: %w", err)
	}
	defer cursor.Close(ctx)
	topics := []model.Topic{}
	if err := cursor.All(ctx, &topics); err != nil {
		return nil, 0, fmt.Errorf("decode topics: %w", err)
	}
	return topics, total, nil
}

func (r *PostRepository) CreateTopic(ctx context.Context, t *model.Topic) error {
	if t.ID.IsZero() {
		t.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.Database().Collection("topics").InsertOne(ctx, t); err != nil {
		return fmt.Errorf("insert topic: %w", err)
	}
	return nil
}

func (r *PostRepository) FindTopic(ctx context.Context, name string) (*model.Topic, error) {
	var t model.Topic
	if err := r.coll.Database().Collection("topics").FindOne(ctx, bson.M{"name": name}).Decode(&t); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find topic: %w", err)
	}
	return &t, nil
}

func cloneFilter(f bson.M) bson.M {
	out := make(bson.M, len(f))
	for k, v := range f {
		out[k] = v
	}
	return out
}

func now() primitive.DateTime {
	return primitive.NewDateTimeFromTime(time.Now())
}
