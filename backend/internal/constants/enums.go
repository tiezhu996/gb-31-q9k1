package constants

// 共享枚举/常量集中定义，模型/DTO/service 状态机/handler 校验/前端筛选均引用本文件。
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) Valid() bool { return r == RoleUser || r == RoleAdmin }

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusBanned UserStatus = "banned"
)

type PetSpecies string

const (
	PetSpeciesDog   PetSpecies = "dog"
	PetSpeciesCat   PetSpecies = "cat"
	PetSpeciesOther PetSpecies = "other"
)

type PetGender string

const (
	PetGenderMale   PetGender = "male"
	PetGenderFemale PetGender = "female"
)

type PetStatus string

const (
	PetStatusActive PetStatus = "active"
	PetStatusHidden PetStatus = "hidden"
)

type PostType string

const (
	PostTypeImage PostType = "image"
	PostTypeVideo PostType = "video"
)

type PostStatus string

const (
	PostStatusPending  PostStatus = "pending"
	PostStatusApproved PostStatus = "approved"
	PostStatusRejected PostStatus = "rejected"
)

type MeetupStatus string

const (
	MeetupStatusOpen      MeetupStatus = "open"
	MeetupStatusFull      MeetupStatus = "full"
	MeetupStatusCancelled MeetupStatus = "cancelled"
	MeetupStatusCompleted MeetupStatus = "completed"
)

type MeetupJoinStatus string

const (
	MeetupJoinJoined    MeetupJoinStatus = "joined"
	MeetupJoinCancelled MeetupJoinStatus = "cancelled"
)

type InteractionType string

const (
	InteractionLike     InteractionType = "like"
	InteractionFavorite InteractionType = "favorite"
	InteractionForward  InteractionType = "forward"
)

type ChatMessageType string

const (
	ChatMessageText    ChatMessageType = "text"
	ChatMessageImage   ChatMessageType = "image"
	ChatMessageSticker ChatMessageType = "sticker"
)

type TopicStatus string

const (
	TopicStatusActive   TopicStatus = "active"
	TopicStatusArchived TopicStatus = "archived"
)
