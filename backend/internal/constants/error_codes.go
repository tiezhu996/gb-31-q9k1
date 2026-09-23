package constants

// 全局错误码集中维护，业务模块不得自造错误码。
const (
	CodeSuccess           = 0
	CodeBadRequest        = 40000
	CodeUnauthorized      = 40100
	CodeTokenExpired      = 40101
	CodeForbidden         = 40300
	CodeRoleForbidden     = 40301
	CodeNotFound          = 40400
	CodeConflict          = 40900
	CodeValidationFailed  = 42200
	CodeContentRejected   = 42201
	CodeRateLimited       = 42900
	CodeInternal          = 50000
	CodeDBError           = 50001
	CodeRedisError        = 50002
	CodeMinIOError        = 50003
	CodeWSUpgradeError    = 50004
	CodeUserExists        = 40901
	CodeWrongPassword     = 40102
	CodeUserBanned        = 40302
	CodePetNotOwned       = 40303
	CodeMeetupFull        = 40902
	CodeMeetupJoined      = 40903
	CodeInteractionExists = 40904
	CodeMeetupConflict    = 40905
	CodeInvalidMediaType  = 42202
)
