package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/petsocial/petsocial/internal/model"
)

// AuditRepository 审计日志数据访问层。
type AuditRepository struct {
	coll *mongo.Collection
}

// NewAuditRepository 构造注入。
func NewAuditRepository(db *mongo.Database) *AuditRepository {
	return &AuditRepository{coll: db.Collection("audit_logs")}
}

func (r *AuditRepository) Create(ctx context.Context, a *model.AuditLog) error {
	if a.ID.IsZero() {
		a.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, a); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func (r *AuditRepository) List(ctx context.Context, action, username string, page, pageSize int) ([]model.AuditLog, int64, error) {
	filter := bson.M{}
	if action != "" {
		filter["action"] = action
	}
	if username != "" {
		filter["username"] = username
	}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer cursor.Close(ctx)
	logs := []model.AuditLog{}
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, fmt.Errorf("decode audit logs: %w", err)
	}
	return logs, total, nil
}
