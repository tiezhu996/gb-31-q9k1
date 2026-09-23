package router

// 本文件确保各实体路由注册函数均被引用，避免误删 import。
var _ = RegisterUserPublic
var _ = RegisterUserAuth
var _ = RegisterPetPublic
var _ = RegisterPetAuth
var _ = RegisterPostPublic
var _ = RegisterPostAuth
var _ = RegisterMeetupPublic
var _ = RegisterMeetupAuth
var _ = RegisterChatAuth
var _ = RegisterFeedPublic
var _ = RegisterMediaAuth
var _ = RegisterAdmin
