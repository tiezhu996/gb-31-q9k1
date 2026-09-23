package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/petsocial/petsocial/internal/config"
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/util"
)

func newTestPostService() *PostService {
	cfg := &config.Config{SensitiveWords: "违禁词,赌博,暴力,色情,诈骗"}
	return &PostService{cfg: cfg, logger: slog.New(slog.NewTextHandler(os.Stderr, nil))}
}

func TestCheckSensitive(t *testing.T) {
	svc := newTestPostService()
	tests := []struct {
		name    string
		content string
		expect  string
	}{
		{"no sensitive", "今天带柴犬去公园散步，开心", ""},
		{"hit 违禁词", "这条内容包含违禁词", "违禁词"},
		{"hit 赌博", "参与网络赌博害人不浅", "赌博"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.checkSensitive(tt.content); got != tt.expect {
				t.Fatalf("checkSensitive(%q) = %q, want %q", tt.content, got, tt.expect)
			}
		})
	}
}

func TestCreatePostRejectedOnSensitive(t *testing.T) {
	svc := newTestPostService()
	_, err := svc.CreatePost(context.Background(), objectIDZero(), dto.CreatePostRequest{
		Content: "这条动态包含赌博内容",
		Type:    constants.PostTypeImage,
	})
	if err == nil {
		t.Fatal("expected error for sensitive content")
	}
	var appErr *util.AppError
	if !errorsAs(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != constants.CodeContentRejected {
		t.Fatalf("expected code %d, got %d", constants.CodeContentRejected, appErr.Code)
	}
}
