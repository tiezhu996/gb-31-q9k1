package repository

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

func TestPostRepositoryListAndStatus(t *testing.T) {
	db := testDB(t)
	repo := NewPostRepository(db)
	ctx := context.Background()
	cleanCollection(t, db, "posts")
	cleanCollection(t, db, "interactions")

	author := primitive.NewObjectID()
	post := &model.Post{
		AuthorID:  author,
		Content:   "测试动态内容",
		Type:      constants.PostTypeImage,
		Status:    constants.PostStatusApproved,
		Topics:    []string{"柴犬日常"},
		City:      "上海",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("create post: %v", err)
	}
	posts, total, err := repo.List(ctx, map[string]interface{}{}, 1, 10)
	if err != nil {
		t.Fatalf("list posts: %v", err)
	}
	if total != 1 || len(posts) != 1 {
		t.Fatalf("total=%d len=%d", total, len(posts))
	}
	byTopic, total, _ := repo.ListByTopic(ctx, "柴犬日常", 1, 10)
	if total != 1 || byTopic[0].ID != post.ID {
		t.Fatalf("topic list mismatch total=%d", total)
	}
	if err := repo.UpdateStatus(ctx, post.ID, constants.PostStatusRejected, "违规"); err != nil {
		t.Fatalf("update status: %v", err)
	}
	got, _ := repo.FindByID(ctx, post.ID)
	if got.Status != constants.PostStatusRejected || got.RejectReason != "违规" {
		t.Fatalf("status=%s reason=%s", got.Status, got.RejectReason)
	}
}

func TestPostRepositoryInteractionUnique(t *testing.T) {
	db := testDB(t)
	ensureTestIndexes(t, db)
	repo := NewPostRepository(db)
	ctx := context.Background()
	cleanCollection(t, db, "posts")
	cleanCollection(t, db, "interactions")

	post := &model.Post{AuthorID: primitive.NewObjectID(), Content: "x", Type: constants.PostTypeImage, Status: constants.PostStatusApproved, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := repo.Create(ctx, post); err != nil {
		t.Fatalf("create post: %v", err)
	}
	user := primitive.NewObjectID()
	if err := repo.CreateInteraction(ctx, &model.Interaction{PostID: post.ID, UserID: user, Type: constants.InteractionLike, CreatedAt: time.Now()}); err != nil {
		t.Fatalf("create interaction: %v", err)
	}
	if err := repo.CreateInteraction(ctx, &model.Interaction{PostID: post.ID, UserID: user, Type: constants.InteractionLike, CreatedAt: time.Now()}); err == nil {
		t.Fatal("expected duplicate interaction error")
	}
	has, _ := repo.HasInteraction(ctx, post.ID, user, constants.InteractionLike)
	if !has {
		t.Fatal("expected interaction exists")
	}
	if err := repo.UpdateCounts(ctx, post.ID, 1, 0, 0, 0); err != nil {
		t.Fatalf("update counts: %v", err)
	}
	got, _ := repo.FindByID(ctx, post.ID)
	if got.LikeCount != 1 {
		t.Fatalf("like count = %d", got.LikeCount)
	}
}
