package service

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/dto"
	"github.com/petsocial/petsocial/internal/model"
)

// fakeMeetupStore 内存版 meetupDataStore：用互斥锁模拟 MongoDB 锁集合的原子语义。
type fakeMeetupStore struct {
	mu      sync.Mutex
	meetups map[primitive.ObjectID]*model.Meetup
	locks   map[primitive.ObjectID]time.Time
	lockTTL time.Duration
}

func newFakeMeetupStore() *fakeMeetupStore {
	return &fakeMeetupStore{
		meetups: map[primitive.ObjectID]*model.Meetup{},
		locks:   map[primitive.ObjectID]time.Time{},
		lockTTL: 10 * time.Second,
	}
}

func (f *fakeMeetupStore) Create(_ context.Context, m *model.Meetup) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m.ID.IsZero() {
		m.ID = primitive.NewObjectID()
	}
	cp := *m
	f.meetups[m.ID] = &cp
	return nil
}

func (f *fakeMeetupStore) FindByID(_ context.Context, id primitive.ObjectID) (*model.Meetup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.meetups[id]
	if !ok {
		return nil, constants.ErrNotFound
	}
	cp := *m
	cp.Participants = append([]model.MeetupParticipant(nil), m.Participants...)
	return &cp, nil
}

func (f *fakeMeetupStore) Update(_ context.Context, m *model.Meetup) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.meetups[m.ID]; !ok {
		return constants.ErrNotFound
	}
	cp := *m
	f.meetups[m.ID] = &cp
	return nil
}

func (f *fakeMeetupStore) UpdateStatus(_ context.Context, id primitive.ObjectID, status constants.MeetupStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.meetups[id]
	if !ok {
		return constants.ErrNotFound
	}
	m.Status = status
	return nil
}

func (f *fakeMeetupStore) List(_ context.Context, _, _ string, _, _ int) ([]model.Meetup, int64, error) {
	return nil, 0, nil
}

func (f *fakeMeetupStore) ListByCreator(_ context.Context, _ primitive.ObjectID, _, _ int) ([]model.Meetup, int64, error) {
	return nil, 0, nil
}

func (f *fakeMeetupStore) FindConflictingJoined(_ context.Context, userID primitive.ObjectID, start, end time.Time, excludeID primitive.ObjectID) ([]model.Meetup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []model.Meetup{}
	for _, m := range f.meetups {
		if m.ID == excludeID {
			continue
		}
		if m.Status != constants.MeetupStatusOpen && m.Status != constants.MeetupStatusFull {
			continue
		}
		joined := false
		for _, p := range m.Participants {
			if p.UserID == userID && p.Status == constants.MeetupJoinJoined {
				joined = true
				break
			}
		}
		if !joined {
			continue
		}
		mEnd := m.MeetTime.Add(time.Duration(m.DurationMinutes) * time.Minute)
		if model.TimeOverlap(m.MeetTime, mEnd, start, end) {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeMeetupStore) AcquireJoinLock(_ context.Context, userID primitive.ObjectID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	exp, ok := f.locks[userID]
	now := time.Now()
	if !ok || !exp.After(now) {
		f.locks[userID] = now.Add(f.lockTTL)
		return true, nil
	}
	return false, nil
}

func (f *fakeMeetupStore) ReleaseJoinLock(_ context.Context, userID primitive.ObjectID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.locks, userID)
	return nil
}

func testMeetupService(f *fakeMeetupStore) *MeetupService {
	return &MeetupService{repo: f, audit: nil, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}

// seedOverlappingMeetups 创建两场时间重叠的招募中约伴。
func seedOverlappingMeetups(t *testing.T, f *fakeMeetupStore) (creator primitive.ObjectID, idA, idB primitive.ObjectID) {
	t.Helper()
	creator = primitive.NewObjectID()
	start := time.Now().Add(24 * time.Hour).Truncate(time.Minute)
	a := &model.Meetup{
		CreatorID: creator, Title: "上午场 A", City: "上海", Location: "公园",
		MeetTime: start, DurationMinutes: 120, MaxPeople: 10,
		Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{},
	}
	b := &model.Meetup{
		CreatorID: creator, Title: "上午场 B", City: "上海", Location: "江边",
		MeetTime: start.Add(30 * time.Minute), DurationMinutes: 60, MaxPeople: 10,
		Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{},
	}
	if err := f.Create(context.Background(), a); err != nil {
		t.Fatal(err)
	}
	if err := f.Create(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	return creator, a.ID, b.ID
}

// TestJoinSequentialConflict 串行报名两场撞车约伴：第二场被拒，且不留报名记录。
func TestJoinSequentialConflict(t *testing.T) {
	f := newFakeMeetupStore()
	svc := testMeetupService(f)
	_, idA, idB := seedOverlappingMeetups(t, f)
	user := primitive.NewObjectID()
	ctx := context.Background()

	if err := svc.Join(ctx, user, idA, ""); err != nil {
		t.Fatalf("join A: %v", err)
	}
	err := svc.Join(ctx, user, idB, "")
	if err == nil {
		t.Fatal("join B should be rejected due to schedule conflict")
	}
	mb, _ := f.FindByID(ctx, idB)
	for _, p := range mb.Participants {
		if p.UserID == user {
			t.Fatalf("failed join must not leave participant record: %+v", p)
		}
	}
}

// TestJoinConcurrentConflict 两个撞车活动并发报名，恰好一个成功。
func TestJoinConcurrentConflict(t *testing.T) {
	f := newFakeMeetupStore()
	svc := testMeetupService(f)
	_, idA, idB := seedOverlappingMeetups(t, f)
	user := primitive.NewObjectID()
	ctx := context.Background()

	start := make(chan struct{})
	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); <-start; errs[0] = svc.Join(ctx, user, idA, "") }()
	go func() { defer wg.Done(); <-start; errs[1] = svc.Join(ctx, user, idB, "") }()
	close(start)
	wg.Wait()

	successes, failures := 0, 0
	for _, err := range errs {
		if err == nil {
			successes++
		} else {
			failures++
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("want exactly 1 success / 1 failure, got %d / %d (%v)", successes, failures, errs)
	}

	for _, id := range []primitive.ObjectID{idA, idB} {
		m, _ := f.FindByID(ctx, id)
		records := 0
		for _, p := range m.Participants {
			if p.UserID == user {
				records++
			}
		}
		if records > 1 {
			t.Fatalf("meetup %s has %d records for user", id.Hex(), records)
		}
	}
}

// TestJoinConflictReleasedAfterCancel 取消报名后时段释放，可报另一场撞车约伴。
func TestJoinConflictReleasedAfterCancel(t *testing.T) {
	f := newFakeMeetupStore()
	svc := testMeetupService(f)
	_, idA, idB := seedOverlappingMeetups(t, f)
	user := primitive.NewObjectID()
	ctx := context.Background()

	if err := svc.Join(ctx, user, idA, ""); err != nil {
		t.Fatalf("join A: %v", err)
	}
	if err := svc.CancelJoin(ctx, user, idA); err != nil {
		t.Fatalf("cancel A: %v", err)
	}
	if err := svc.Join(ctx, user, idB, ""); err != nil {
		t.Fatalf("join B after cancelling A should succeed: %v", err)
	}
}

// TestJoinBackToBackNotConflict 时间首尾相接（半开区间）不算撞车。
func TestJoinBackToBackNotConflict(t *testing.T) {
	f := newFakeMeetupStore()
	svc := testMeetupService(f)
	creator := primitive.NewObjectID()
	start := time.Now().Add(24 * time.Hour).Truncate(time.Minute)
	a := &model.Meetup{CreatorID: creator, Title: "第一场", MeetTime: start, DurationMinutes: 120, MaxPeople: 10, Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{}}
	b := &model.Meetup{CreatorID: creator, Title: "第二场", MeetTime: start.Add(120 * time.Minute), DurationMinutes: 60, MaxPeople: 10, Status: constants.MeetupStatusOpen, Participants: []model.MeetupParticipant{}}
	ctx := context.Background()
	if err := f.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	if err := f.Create(ctx, b); err != nil {
		t.Fatal(err)
	}
	user := primitive.NewObjectID()
	if err := svc.Join(ctx, user, a.ID, ""); err != nil {
		t.Fatalf("join A: %v", err)
	}
	if err := svc.Join(ctx, user, b.ID, ""); err != nil {
		t.Fatalf("back-to-back join B should succeed: %v", err)
	}
}

// TestJoinCancelledMeetupDoesNotBlock 已取消的约伴不占用时段。
func TestJoinCancelledMeetupDoesNotBlock(t *testing.T) {
	f := newFakeMeetupStore()
	svc := testMeetupService(f)
	_, idA, idB := seedOverlappingMeetups(t, f)
	user := primitive.NewObjectID()
	ctx := context.Background()

	if err := svc.Join(ctx, user, idA, ""); err != nil {
		t.Fatalf("join A: %v", err)
	}
	if err := svc.UpdateStatus(ctx, f.meetups[idA].CreatorID, idA, dto.UpdateMeetupStatusRequest{Status: constants.MeetupStatusCancelled}); err != nil {
		t.Fatalf("cancel meetup A: %v", err)
	}
	if err := svc.Join(ctx, user, idB, ""); err != nil {
		t.Fatalf("join B after meetup A cancelled should succeed: %v", err)
	}
}
