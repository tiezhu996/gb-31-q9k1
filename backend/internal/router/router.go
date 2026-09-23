package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/petsocial/petsocial/internal/config"
	"github.com/petsocial/petsocial/internal/constants"
	"github.com/petsocial/petsocial/internal/handler"
	"github.com/petsocial/petsocial/internal/middleware"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/service"
)

// Deps 路由装配所需的全部依赖。
type Deps struct {
	Cfg       *config.Config
	Logger    *slog.Logger
	Repo      *RepositoryBundle
	Service   *ServiceBundle
	Handler   *HandlerBundle
	Redis     *redis.Client
	JWTSecret string
}

// RepositoryBundle 仓储聚合。
type RepositoryBundle struct {
	User   *repository.UserRepository
	Pet    *repository.PetRepository
	Post   *repository.PostRepository
	Meetup *repository.MeetupRepository
	Chat   *repository.ChatRepository
	Audit  *repository.AuditRepository
}

// ServiceBundle 服务聚合。
type ServiceBundle struct {
	User   *service.UserService
	Pet    *service.PetService
	Post   *service.PostService
	Meetup *service.MeetupService
	Chat   *service.ChatService
	Audit  *service.AuditService
	Feed   *service.FeedService
	Media  *service.MediaService
}

// HandlerBundle 接口聚合。
type HandlerBundle struct {
	User   *handler.UserHandler
	Pet    *handler.PetHandler
	Post   *handler.PostHandler
	Meetup *handler.MeetupHandler
	Chat   *handler.ChatHandler
	Audit  *handler.AuditHandler
	Feed   *handler.FeedHandler
	Media  *handler.MediaHandler
}

// New 组装全部路由与中间件。
func New(deps *Deps) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(deps.Logger))
	r.Use(middleware.Recovery(deps.Logger))
	r.Use(middleware.CORS(deps.Cfg.CORSOrigins))
	r.Use(middleware.ErrorHandler(deps.Logger))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.RateLimit(deps.Redis, deps.Cfg.RateLimit))

	public := api.Group("")
	public.Use(middleware.OptionalAuth(deps.JWTSecret))
	RegisterUserPublic(public, deps.Handler.User)
	RegisterPostPublic(public, deps.Handler.Post)
	RegisterFeedPublic(public, deps.Handler.Feed)
	RegisterPetPublic(public, deps.Handler.Pet)
	RegisterMeetupPublic(public, deps.Handler.Meetup)

	auth := api.Group("")
	auth.Use(middleware.Auth(deps.JWTSecret))
	RegisterUserAuth(auth, deps.Handler.User)
	RegisterPetAuth(auth, deps.Handler.Pet)
	RegisterPostAuth(auth, deps.Handler.Post)
	RegisterMeetupAuth(auth, deps.Handler.Meetup)
	RegisterChatAuth(auth, deps.Handler.Chat)
	RegisterMediaAuth(auth, deps.Handler.Media)

	admin := api.Group("")
	admin.Use(middleware.Auth(deps.JWTSecret))
	admin.Use(middleware.RequireRole(constants.RoleAdmin))
	RegisterAdmin(admin, deps.Handler.Post, deps.Handler.Audit)

	return r
}
