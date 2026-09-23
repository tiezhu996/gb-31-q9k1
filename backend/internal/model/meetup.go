package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/petsocial/petsocial/internal/constants"
)

// MeetupParticipant 约伴参与者。
type MeetupParticipant struct {
	UserID   primitive.ObjectID         `bson:"user_id" json:"user_id"`
	JoinedAt time.Time                  `bson:"joined_at" json:"joined_at"`
	Status   constants.MeetupJoinStatus `bson:"status" json:"status"`
}

// Meetup 同城遛狗搭子约伴帖：报名并发安全依赖 meetup_join_locks 用户级锁串行化。
type Meetup struct {
	ID              primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	CreatorID       primitive.ObjectID     `bson:"creator_id" json:"creator_id"`
	Title           string                 `bson:"title" json:"title"`
	Description     string                 `bson:"description" json:"description"`
	City            string                 `bson:"city" json:"city"`
	Location        string                 `bson:"location" json:"location"`
	MeetTime        time.Time              `bson:"meet_time" json:"meet_time"`
	DurationMinutes int                    `bson:"duration_minutes" json:"duration_minutes"`
	MaxPeople       int                    `bson:"max_people" json:"max_people"`
	Status          constants.MeetupStatus `bson:"status" json:"status"`
	Participants    []MeetupParticipant    `bson:"participants" json:"participants"`
	CreatedAt       time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `bson:"updated_at" json:"updated_at"`
}

// EndTime 约伴结束时间 = 开始时间 + 时长。
func (m *Meetup) EndTime() time.Time {
	return m.MeetTime.Add(time.Duration(m.DurationMinutes) * time.Minute)
}

// TimeOverlap 两个半开区间 [start, end) 是否重叠：端点相接不算撞车。
func TimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
