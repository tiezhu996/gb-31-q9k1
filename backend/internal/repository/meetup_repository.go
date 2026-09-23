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

// MeetupRepository 约伴帖数据访问层。
type MeetupRepository struct {
	coll *mongo.Collection
}

// NewMeetupRepository 构造注入。
func NewMeetupRepository(db *mongo.Database) *MeetupRepository {
	return &MeetupRepository{coll: db.Collection("meetups")}
}

func (r *MeetupRepository) Create(ctx context.Context, m *model.Meetup) error {
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, m); err != nil {
		return fmt.Errorf("insert meetup: %w", err)
	}
	return nil
}

func (r *MeetupRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Meetup, error) {
	var m model.Meetup
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&m); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find meetup by id: %w", err)
	}
	return &m, nil
}

func (r *MeetupRepository) Update(ctx context.Context, m *model.Meetup) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": m.ID}, m)
	if err != nil {
		return fmt.Errorf("update meetup: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *MeetupRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status constants.MeetupStatus) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status, "updated_at": now()}})
	if err != nil {
		return fmt.Errorf("update meetup status: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *MeetupRepository) List(ctx context.Context, city, status string, page, pageSize int) ([]model.Meetup, int64, error) {
	filter := bson.M{}
	if city != "" {
		filter["city"] = city
	}
	if status != "" {
		filter["status"] = status
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count meetups: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"meet_time": 1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list meetups: %w", err)
	}
	defer cursor.Close(ctx)
	meetups := []model.Meetup{}
	if err := cursor.All(ctx, &meetups); err != nil {
		return nil, 0, fmt.Errorf("decode meetups: %w", err)
	}
	return meetups, total, nil
}

func (r *MeetupRepository) ListByCreator(ctx context.Context, creatorID primitive.ObjectID, page, pageSize int) ([]model.Meetup, int64, error) {
	filter := bson.M{"creator_id": creatorID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count meetups by creator: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list meetups by creator: %w", err)
	}
	defer cursor.Close(ctx)
	meetups := []model.Meetup{}
	if err := cursor.All(ctx, &meetups); err != nil {
		return nil, 0, fmt.Errorf("decode meetups by creator: %w", err)
	}
	return meetups, total, nil
}
