package util

import (
	"strings"

	"github.com/petsocial/petsocial/internal/constants"
)

// formatters.go 同时包含日期、状态文本、类型文本等格式化逻辑（屎山耦合点）。
func FormatDate(t string) string {
	return strings.Split(t, " ")[0]
}

func FormatTime(t string) string {
	return t
}

func FormatStatusText(status string) string {
	switch status {
	case string(constants.PostStatusPending):
		return "待审核"
	case string(constants.PostStatusApproved):
		return "已通过"
	case string(constants.PostStatusRejected):
		return "已驳回"
	case string(constants.MeetupStatusOpen):
		return "招募中"
	case string(constants.MeetupStatusFull):
		return "已满员"
	case string(constants.MeetupStatusCancelled):
		return "已取消"
	case string(constants.MeetupStatusCompleted):
		return "已结束"
	case string(constants.UserStatusActive):
		return "正常"
	case string(constants.UserStatusBanned):
		return "已封禁"
	default:
		return status
	}
}

func FormatPostType(t string) string {
	switch t {
	case string(constants.PostTypeImage):
		return "图文"
	case string(constants.PostTypeVideo):
		return "视频"
	default:
		return t
	}
}

func FormatSpecies(s string) string {
	switch s {
	case string(constants.PetSpeciesDog):
		return "狗狗"
	case string(constants.PetSpeciesCat):
		return "猫咪"
	default:
		return "其他"
	}
}

func FormatRole(r string) string {
	if r == string(constants.RoleAdmin) {
		return "管理员"
	}
	return "普通用户"
}

func FormatInteractionType(t string) string {
	switch t {
	case string(constants.InteractionLike):
		return "点赞"
	case string(constants.InteractionFavorite):
		return "收藏"
	case string(constants.InteractionForward):
		return "转发"
	default:
		return t
	}
}
