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

// PetRepository 宠物数据访问层。
type PetRepository struct {
	coll *mongo.Collection
}

// NewPetRepository 构造注入。
func NewPetRepository(db *mongo.Database) *PetRepository {
	return &PetRepository{coll: db.Collection("pets")}
}

func (r *PetRepository) Create(ctx context.Context, p *model.Pet) error {
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.InsertOne(ctx, p); err != nil {
		return fmt.Errorf("insert pet: %w", err)
	}
	return nil
}

func (r *PetRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Pet, error) {
	var p model.Pet
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("find pet by id: %w", err)
	}
	return &p, nil
}

func (r *PetRepository) Update(ctx context.Context, p *model.Pet) error {
	res, err := r.coll.ReplaceOne(ctx, bson.M{"_id": p.ID}, p)
	if err != nil {
		return fmt.Errorf("update pet: %w", err)
	}
	if res.MatchedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PetRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("delete pet: %w", err)
	}
	if res.DeletedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PetRepository) ListByOwner(ctx context.Context, ownerID primitive.ObjectID, page, pageSize int) ([]model.Pet, int64, error) {
	filter := bson.M{"owner_id": ownerID}
	return r.list(ctx, filter, page, pageSize)
}

func (r *PetRepository) List(ctx context.Context, species, city string, page, pageSize int) ([]model.Pet, int64, error) {
	filter := bson.M{"status": constants.PetStatusActive}
	if species != "" {
		filter["species"] = species
	}
	if city != "" {
		filter["city"] = city
	}
	return r.list(ctx, filter, page, pageSize)
}

func (r *PetRepository) list(ctx context.Context, filter bson.M, page, pageSize int) ([]model.Pet, int64, error) {
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count pets: %w", err)
	}
	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list pets: %w", err)
	}
	defer cursor.Close(ctx)
	pets := []model.Pet{}
	if err := cursor.All(ctx, &pets); err != nil {
		return nil, 0, fmt.Errorf("decode pets: %w", err)
	}
	return pets, total, nil
}

// FollowPet 关注宠物（唯一索引防重）。
func (r *PetRepository) FollowPet(ctx context.Context, f *model.PetFollow) error {
	if f.ID.IsZero() {
		f.ID = primitive.NewObjectID()
	}
	if _, err := r.coll.Database().Collection("pet_follows").InsertOne(ctx, f); err != nil {
		return fmt.Errorf("insert pet follow: %w", err)
	}
	return nil
}

func (r *PetRepository) UnfollowPet(ctx context.Context, petID, userID primitive.ObjectID) error {
	res, err := r.coll.Database().Collection("pet_follows").DeleteOne(ctx, bson.M{"pet_id": petID, "user_id": userID})
	if err != nil {
		return fmt.Errorf("delete pet follow: %w", err)
	}
	if res.DeletedCount == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *PetRepository) UpdateFollowerCount(ctx context.Context, petID primitive.ObjectID, inc int64) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": petID}, bson.M{"$inc": bson.M{"follower_count": inc}})
	if err != nil {
		return fmt.Errorf("update pet follower count: %w", err)
	}
	return nil
}
