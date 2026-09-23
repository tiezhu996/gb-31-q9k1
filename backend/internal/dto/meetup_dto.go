package dto

import (
	"time"

	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateMeetupRequest 发布约伴帖入参。
type CreateMeetupRequest struct {
	Title       string `json:"title" binding:"required,min=1,max=64"`
	Description string `json:"description" binding:"max=500"`
	City        string `json:"city" binding:"required,max=32"`
	Location    string `json:"location" binding:"required,max=128"`
	MeetTime    string `json:"meet_time" binding:"required"`
	MaxPeople   int    `json:"max_people" binding:"required,min=2,max=50"`
}

// UpdateMeetupStatusRequest 约伴状态流转入参。
type UpdateMeetupStatusRequest struct {
	Status constants.MeetupStatus `json:"status" binding:"required,oneof=cancelled completed"`
}

// MeetupResponse 约伴出参。
type MeetupResponse struct {
	ID           string                 `json:"id"`
	CreatorID    string                 `json:"creator_id"`
	CreatorName  string                 `json:"creator_name"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	City         string                 `json:"city"`
	Location     string                 `json:"location"`
	MeetTime     string                 `json:"meet_time"`
	MaxPeople    int                    `json:"max_people"`
	Status       constants.MeetupStatus `json:"status"`
	JoinedCount  int                    `json:"joined_count"`
	Joined       bool                   `json:"joined"`
	Participants []MeetupParticipantDTO `json:"participants"`
	CreatedAt    string                 `json:"created_at"`
}

// MeetupParticipantDTO 参与者出参。
type MeetupParticipantDTO struct {
	UserID   string                     `json:"user_id"`
	Username string                     `json:"username"`
	JoinedAt string                     `json:"joined_at"`
	Status   constants.MeetupJoinStatus `json:"status"`
}

// ToMeetupResponse 模型转出参。
func ToMeetupResponse(m *model.Meetup, creatorName string, currentUserID primitive.ObjectID) MeetupResponse {
	joined := false
	joinedCount := 0
	participants := make([]MeetupParticipantDTO, 0, len(m.Participants))
	for _, p := range m.Participants {
		if p.Status == constants.MeetupJoinJoined {
			joinedCount++
		}
		if p.UserID == currentUserID && p.Status == constants.MeetupJoinJoined {
			joined = true
		}
		participants = append(participants, MeetupParticipantDTO{
			UserID:   p.UserID.Hex(),
			JoinedAt: p.JoinedAt.Format("2006-01-02 15:04:05"),
			Status:   p.Status,
		})
	}
	return MeetupResponse{
		ID:           m.ID.Hex(),
		CreatorID:    m.CreatorID.Hex(),
		CreatorName:  creatorName,
		Title:        m.Title,
		Description:  m.Description,
		City:         m.City,
		Location:     m.Location,
		MeetTime:     m.MeetTime.Format("2006-01-02 15:04"),
		MaxPeople:    m.MaxPeople,
		Status:       m.Status,
		JoinedCount:  joinedCount,
		Joined:       joined,
		Participants: participants,
		CreatedAt:    m.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ParseMeetTime 解析约伴时间。
func ParseMeetTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04", s)
}
