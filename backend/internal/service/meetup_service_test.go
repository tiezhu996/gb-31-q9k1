package service

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
)

func TestEnsureNoDuplicate(t *testing.T) {
	svc := &MeetupService{}
	u1 := primitive.NewObjectID()
	u2 := primitive.NewObjectID()
	meetup := &model.Meetup{
		Participants: []model.MeetupParticipant{
			{UserID: u1, Status: constants.MeetupJoinJoined},
		},
	}
	tests := []struct {
		name   string
		userID primitive.ObjectID
		expect bool
	}{
		{"not joined yet", u2, true},
		{"already joined", u1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.ensureNoDuplicate(meetup, tt.userID); got != tt.expect {
				t.Fatalf("ensureNoDuplicate = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestIsFull(t *testing.T) {
	svc := &MeetupService{}
	u1 := primitive.NewObjectID()
	meetup := &model.Meetup{MaxPeople: 2, Participants: []model.MeetupParticipant{
		{UserID: u1, Status: constants.MeetupJoinJoined},
	}}
	if svc.isFull(meetup) {
		t.Fatal("expected not full with 1/2")
	}
	u2 := primitive.NewObjectID()
	meetup.Participants = append(meetup.Participants, model.MeetupParticipant{UserID: u2, Status: constants.MeetupJoinJoined})
	if !svc.isFull(meetup) {
		t.Fatal("expected full with 2/2")
	}
}

func TestValidateDuration(t *testing.T) {
	for _, ok := range []int{30, 120, 480} {
		if err := ValidateDuration(ok); err != nil {
			t.Fatalf("ValidateDuration(%d) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []int{0, 29, 481, -120} {
		if err := ValidateDuration(bad); err == nil {
			t.Fatalf("ValidateDuration(%d) = nil, want error", bad)
		}
	}
}

func TestConflictMessage(t *testing.T) {
	meetup := &model.Meetup{
		Title:    "周六世纪公园遛狗局",
		MeetTime: time.Date(2026, 9, 26, 10, 0, 0, 0, time.Local),
	}
	msg := conflictMessage(meetup)
	if !strings.Contains(msg, "周六世纪公园遛狗局") {
		t.Fatalf("conflict message missing title: %s", msg)
	}
	if !strings.Contains(msg, "2026-09-26 10:00") {
		t.Fatalf("conflict message missing meet time: %s", msg)
	}
}
