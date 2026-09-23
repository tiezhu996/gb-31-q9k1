package repository

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

func TestUserRepositoryCRUD(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username:     "repo_tester",
		PasswordHash: "hash",
		Nickname:     "测试用户",
		Role:         constants.RoleUser,
		Status:       constants.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := repo.FindByUsername(ctx, "repo_tester")
	if err != nil {
		t.Fatalf("find by username: %v", err)
	}
	if got.Nickname != "测试用户" {
		t.Fatalf("nickname = %q", got.Nickname)
	}
	got2, err := repo.FindByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got2.ID != user.ID {
		t.Fatalf("id mismatch")
	}
	if err := repo.UpdateCounts(ctx, user.ID, 1, 2); err != nil {
		t.Fatalf("update counts: %v", err)
	}
	got3, _ := repo.FindByID(ctx, user.ID)
	if got3.FollowCount != 1 || got3.FollowerCount != 2 {
		t.Fatalf("counts = %d/%d", got3.FollowCount, got3.FollowerCount)
	}
	if _, err := repo.FindByUsername(ctx, "nobody"); err != constants.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepositoryFollowUnique(t *testing.T) {
	db := testDB(t)
	ensureTestIndexes(t, db)
	repo := NewUserRepository(db)
	ctx := context.Background()
	cleanCollection(t, db, "follows")

	f1 := primitive.NewObjectID()
	f2 := primitive.NewObjectID()
	follow := &model.Follow{FollowerID: f1, FolloweeID: f2, CreatedAt: time.Now()}
	if err := repo.CreateFollow(ctx, follow); err != nil {
		t.Fatalf("create follow: %v", err)
	}
	// 重复关注应命中唯一索引报错
	if err := repo.CreateFollow(ctx, &model.Follow{FollowerID: f1, FolloweeID: f2, CreatedAt: time.Now()}); err == nil {
		t.Fatal("expected duplicate key error")
	}
	ok, err := repo.IsFollowing(ctx, f1, f2)
	if err != nil || !ok {
		t.Fatalf("is following = %v, %v", ok, err)
	}
	if err := repo.DeleteFollow(ctx, f1, f2); err != nil {
		t.Fatalf("delete follow: %v", err)
	}
	ok, _ = repo.IsFollowing(ctx, f1, f2)
	if ok {
		t.Fatal("expected not following after delete")
	}
}
