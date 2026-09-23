package service

import (
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

func TestNormalizeDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    int
		wantErr bool
	}{
		{"empty defaults to 120", 0, constants.MeetupDurationDefault, false},
		{"min boundary 30", 30, 30, false},
		{"max boundary 480", 480, 480, false},
		{"below min", 29, 0, true},
		{"above max", 481, 0, true},
		{"negative", -10, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NormalizeDuration(%d) expected error, got %d", tt.input, got)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("NormalizeDuration(%d) = %d, %v; want %d", tt.input, got, err, tt.want)
			}
		})
	}
}

func TestTimeRangesOverlap(t *testing.T) {
	base := time.Date(2026, 9, 23, 10, 0, 0, 0, time.Local)
	at := func(hour, minute int) time.Time {
		return time.Date(2026, 9, 23, hour, minute, 0, 0, time.Local)
	}
	tests := []struct {
		name         string
		start1, end1 time.Time
		start2, end2 time.Time
		want         bool
	}{
		{"full containment", base, at(12, 0), at(10, 30), at(11, 30), true},
		{"partial overlap", base, at(12, 0), at(11, 0), at(13, 0), true},
		{"back-to-back not overlap", base, at(12, 0), at(12, 0), at(14, 0), false},
		{"separate ranges", base, at(11, 0), at(12, 0), at(13, 0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := timeRangesOverlap(tt.start1, tt.end1, tt.start2, tt.end2); got != tt.want {
				t.Fatalf("timeRangesOverlap = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindMeetupConflict(t *testing.T) {
	user := primitive.NewObjectID()
	targetID := primitive.NewObjectID()
	base := time.Date(2026, 9, 23, 10, 0, 0, 0, time.Local)
	mk := func(title string, start time.Time, duration int, status constants.MeetupStatus, joinStatus constants.MeetupJoinStatus) model.Meetup {
		return model.Meetup{
			ID:              primitive.NewObjectID(),
			Title:           title,
			MeetTime:        start,
			DurationMinutes: duration,
			Status:          status,
			Participants:    []model.MeetupParticipant{{UserID: user, Status: joinStatus}},
		}
	}
	overlapping := mk("冲突场", base, 120, constants.MeetupStatusOpen, constants.MeetupJoinJoined)
	backToBack := mk("首尾相接", base.Add(2*time.Hour), 60, constants.MeetupStatusOpen, constants.MeetupJoinJoined)
	fullMeetup := mk("已满员也算占用", base, 120, constants.MeetupStatusFull, constants.MeetupJoinJoined)

	t.Run("rejects overlapping joined meetup", func(t *testing.T) {
		active := []model.Meetup{overlapping, backToBack}
		got := findMeetupConflict(active, targetID, base, 120)
		if got == nil || got.Title != "冲突场" {
			t.Fatalf("expected conflict with 冲突场, got %+v", got)
		}
	})
	t.Run("full status still blocks", func(t *testing.T) {
		active := []model.Meetup{fullMeetup}
		if got := findMeetupConflict(active, targetID, base, 120); got == nil {
			t.Fatal("expected full meetup to block")
		}
	})
	t.Run("back to back allowed", func(t *testing.T) {
		active := []model.Meetup{backToBack}
		if got := findMeetupConflict(active, targetID, base, 120); got != nil {
			t.Fatalf("back-to-back should not conflict, got %s", got.Title)
		}
	})
	t.Run("releases after cancel", func(t *testing.T) {
		// 取消报名（joinStatus=cancelled）与约伴取消（status=cancelled）由仓储层
		// ListActiveByParticipant 过滤，不会进入 active 列表，因此 findMeetupConflict
		// 收到的列表为空即代表时间段已释放。
		var active []model.Meetup
		if got := findMeetupConflict(active, targetID, base, 120); got != nil {
			t.Fatalf("empty active list should not conflict, got %s", got.Title)
		}
	})
	t.Run("skips target itself", func(t *testing.T) {
		self := overlapping
		self.ID = targetID
		if got := findMeetupConflict([]model.Meetup{self}, targetID, base, 120); got != nil {
			t.Fatalf("target meetup must be skipped, got %s", got.Title)
		}
	})
	t.Run("legacy meetup without duration defaults to 120", func(t *testing.T) {
		legacy := mk("老数据无时长", base.Add(-30*time.Minute), 0, constants.MeetupStatusOpen, constants.MeetupJoinJoined)
		if got := findMeetupConflict([]model.Meetup{legacy}, targetID, base, 60); got == nil {
			t.Fatal("legacy meetup duration should default to 120 and conflict")
		}
	})
}
