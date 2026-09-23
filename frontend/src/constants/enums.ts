// 与后端 internal/constants/enums.go 对应的共享枚举
export const ROLE = {
  USER: 'user',
  ADMIN: 'admin',
} as const;

export const POST_TYPE = {
  IMAGE: 'image',
  VIDEO: 'video',
} as const;

export const POST_STATUS = {
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
} as const;

export const MEETUP_STATUS = {
  OPEN: 'open',
  FULL: 'full',
  CANCELLED: 'cancelled',
  COMPLETED: 'completed',
} as const;

export const MEETUP_JOIN_STATUS = {
  JOINED: 'joined',
  CANCELLED: 'cancelled',
} as const;

export const INTERACTION_TYPE = {
  LIKE: 'like',
  FAVORITE: 'favorite',
  FORWARD: 'forward',
} as const;

export const PET_SPECIES = {
  DOG: 'dog',
  CAT: 'cat',
  OTHER: 'other',
} as const;

export const PET_GENDER = {
  MALE: 'male',
  FEMALE: 'female',
} as const;

export const USER_STATUS = {
  ACTIVE: 'active',
  BANNED: 'banned',
} as const;

export const CHAT_MESSAGE_TYPE = {
  TEXT: 'text',
  IMAGE: 'image',
  STICKER: 'sticker',
} as const;

export const TOPIC_STATUS = {
  ACTIVE: 'active',
  ARCHIVED: 'archived',
} as const;

export const ROLE_LABEL: Record<string, string> = { user: '普通用户', admin: '管理员' };
export const POST_STATUS_LABEL: Record<string, string> = { pending: '待审核', approved: '已通过', rejected: '已驳回' };
export const MEETUP_STATUS_LABEL: Record<string, string> = { open: '招募中', full: '已满员', cancelled: '已取消', completed: '已结束' };
export const POST_TYPE_LABEL: Record<string, string> = { image: '图文', video: '视频' };
export const PET_SPECIES_LABEL: Record<string, string> = { dog: '狗狗', cat: '猫咪', other: '其他' };
export const USER_STATUS_LABEL: Record<string, string> = { active: '正常', banned: '已封禁' };
export const INTERACTION_TYPE_LABEL: Record<string, string> = { like: '点赞', favorite: '收藏', forward: '转发' };
export const CHAT_MESSAGE_TYPE_LABEL: Record<string, string> = { text: '文本', image: '图片', sticker: '贴纸' };
