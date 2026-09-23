package constants

import "errors"

// 哨兵错误集中定义，仓储层/服务层用 errors.Is 判断。
var (
	ErrNotFound         = errors.New("resource not found")
	ErrInvalidID        = errors.New("invalid object id")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("conflict")
	ErrUserExists       = errors.New("user already exists")
	ErrWrongPassword    = errors.New("wrong password")
	ErrUserBanned       = errors.New("user banned")
	ErrContentRejected  = errors.New("content rejected")
	ErrMeetupFull       = errors.New("meetup full")
	ErrMeetupJoined     = errors.New("already joined")
	ErrMeetupConflict   = errors.New("meetup time conflict")
	ErrInteractionDup   = errors.New("interaction duplicate")
	ErrPetNotOwned      = errors.New("pet not owned")
	ErrRateLimited      = errors.New("rate limited")
	ErrValidation       = errors.New("validation failed")
	ErrInvalidMediaType = errors.New("invalid media type")
	ErrNotImplemented   = errors.New("not implemented")
)
