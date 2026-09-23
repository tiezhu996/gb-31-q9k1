package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

// joinLockTTL 用户报名锁最长持有时间，持有者异常退出后由后续请求抢锁回收。
const joinLockTTL = 10 * time.Second

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

// FindConflictingJoined 查询该用户已报名（未取消）且与给定时间区间重叠的约伴。
// 仅统计仍在招募/已满员（未取消、未结束）的约伴；半开区间 [start, end) 相接不算冲突。
// excludeID 用于排除当前正在报名的那场（防御性，正常情况下本人尚未在其名单中）。
func (r *MeetupRepository) FindConflictingJoined(
	ctx context.Context,
	userID primitive.ObjectID,
	start, end time.Time,
	excludeID primitive.ObjectID,
) ([]model.Meetup, error) {
	filter := bson.M{
		"_id":    bson.M{"$ne": excludeID},
		"status": bson.M{"$in": []constants.MeetupStatus{constants.MeetupStatusOpen, constants.MeetupStatusFull}},
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userID,
				"status":  constants.MeetupJoinJoined,
			},
		},
		"$expr": bson.M{
			"$and": bson.A{
				bson.M{"$lt": bson.A{"$meet_time", end}},
				bson.M{"$gt": bson.A{
					bson.M{"$add": bson.A{"$meet_time", bson.M{"$multiply": bson.A{"$duration_minutes", 60000}}}},
					start,
				}},
			},
		},
	}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find conflicting meetups: %w", err)
	}
	defer cursor.Close(ctx)
	meetups := []model.Meetup{}
	if err := cursor.All(ctx, &meetups); err != nil {
		return nil, fmt.Errorf("decode conflicting meetups: %w", err)
	}
	return meetups, nil
}

// AcquireJoinLock 获取用户级报名锁（_id 唯一索引保证跨并发原子）。
// 首次 InsertOne 成功即获锁；锁已存在但超过 TTL 时抢占回收，否则返回 false 让调用方重试。
func (r *MeetupRepository) AcquireJoinLock(ctx context.Context, userID primitive.ObjectID) (bool, error) {
	now := time.Now()
	_, err := r.coll.Database().Collection("meetup_join_locks").InsertOne(ctx, bson.M{
		"_id":        userID,
		"locked_at":  now,
		"expires_at": now.Add(joinLockTTL),
	})
	if err == nil {
		return true, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return false, fmt.Errorf("acquire join lock: %w", err)
	}
	res, err := r.coll.Database().Collection("meetup_join_locks").UpdateOne(ctx,
		bson.M{"_id": userID, "expires_at": bson.M{"$lte": now}},
		bson.M{"$set": bson.M{"locked_at": now, "expires_at": now.Add(joinLockTTL)}},
	)
	if err != nil {
		return false, fmt.Errorf("steal expired join lock: %w", err)
	}
	return res.ModifiedCount > 0, nil
}

// ReleaseJoinLock 释放用户级报名锁。
func (r *MeetupRepository) ReleaseJoinLock(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.coll.Database().Collection("meetup_join_locks").DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return fmt.Errorf("release join lock: %w", err)
	}
	return nil
}
