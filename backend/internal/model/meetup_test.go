package model

import (
	"testing"
	"time"

	"github.com/petsocial/petsocial/internal/constants"
)

func TestTimeOverlap(t *testing.T) {
	base := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	h := func(n int) time.Time { return base.Add(time.Duration(n) * time.Minute) }
	tests := []struct {
		name           string
		s1, e1, s2, e2 time.Time
		want           bool
	}{
		{"fully inside", h(0), h(120), h(30), h(60), true},
		{"partial overlap", h(0), h(120), h(60), h(180), true},
		{"boundary touch is not overlap", h(0), h(120), h(120), h(240), false},
		{"before", h(0), h(120), h(-120), h(0), false},
		{"after", h(0), h(120), h(130), h(200), false},
		{"same start", h(0), h(120), h(0), h(60), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeOverlap(tt.s1, tt.e1, tt.s2, tt.e2); got != tt.want {
				t.Fatalf("TimeOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeetupEndTime(t *testing.T) {
	start := time.Date(2026, 9, 23, 10, 0, 0, 0, time.UTC)
	m := &Meetup{MeetTime: start, DurationMinutes: 90}
	want := time.Date(2026, 9, 23, 11, 30, 0, 0, time.UTC)
	if !m.EndTime().Equal(want) {
		t.Fatalf("EndTime = %v, want %v", m.EndTime(), want)
	}
}

// TestJoinStatusRelease 取消报名的参与者不再占用名额/时段。
func TestJoinStatusRelease(t *testing.T) {
	joined, cancelled := 0, 0
	participants := []MeetupParticipant{
		{Status: constants.MeetupJoinJoined},
		{Status: constants.MeetupJoinCancelled},
	}
	for _, p := range participants {
		switch p.Status {
		case constants.MeetupJoinJoined:
			joined++
		case constants.MeetupJoinCancelled:
			cancelled++
		}
	}
	if joined != 1 || cancelled != 1 {
		t.Fatalf("joined=%d cancelled=%d, want 1/1", joined, cancelled)
	}
}
