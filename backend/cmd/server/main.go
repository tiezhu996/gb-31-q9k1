package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/petsocial/petsocial/internal/config"
	"github.com/petsocial/petsocial/internal/database"
	"github.com/petsocial/petsocial/internal/handler"
	"github.com/petsocial/petsocial/internal/repository"
	"github.com/petsocial/petsocial/internal/router"
	"github.com/petsocial/petsocial/internal/service"
	"github.com/petsocial/petsocial/internal/util"
	minioclient "github.com/petsocial/petsocial/pkg/minioclient"
	"github.com/petsocial/petsocial/pkg/redisclient"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	logger := util.NewLogger(cfg.AppEnv)
	logger.Info(fmt.Sprintf(logServerStart(cfg)))
	util.InitValidator()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.MongoURI, cfg.DBName, logger)
	if err != nil {
		logger.Error("database connect failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close(context.Background()) }()

	rdb, err := redisclient.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		logger.Error("redis connect failed", "error", err)
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf(redisConnected(cfg)))

	mc, err := minioclient.NewMinIO(ctx, cfg.MinIOEndpoint, cfg.MinIOAccess, cfg.MinIOSecret, cfg.MinIOBucket, cfg.MediaPublicURL, false)
	if err != nil {
		logger.Warn("minio connect failed, media upload disabled", "error", err)
		mc = nil
	}

	if err := database.Seed(ctx, db.Mongo, logger); err != nil {
		logger.Error("seed failed", "error", err)
	}

	mongo := db.Mongo
	userRepo := repository.NewUserRepository(mongo)
	petRepo := repository.NewPetRepository(mongo)
	postRepo := repository.NewPostRepository(mongo)
	meetupRepo := repository.NewMeetupRepository(mongo)
	chatRepo := repository.NewChatRepository(mongo)
	auditRepo := repository.NewAuditRepository(mongo)

	auditSvc := service.NewAuditService(auditRepo, logger)
	userSvc := service.NewUserService(userRepo, auditSvc, logger, cfg.JWTSecret, cfg.JWTExpireHour)
	petSvc := service.NewPetService(petRepo, auditSvc, logger)
	postSvc := service.NewPostService(postRepo, userRepo, auditSvc, logger, cfg)
	meetupSvc := service.NewMeetupService(meetupRepo, auditSvc, logger)
	chatSvc := service.NewChatService(chatRepo, userRepo, auditSvc, logger)
	feedSvc := service.NewFeedService(postRepo, userRepo, logger)
	var mediaSvc *service.MediaService
	if mc != nil {
		mediaSvc = service.NewMediaService(mc, auditSvc, logger)
	}

	userH := handler.NewUserHandler(userSvc, logger)
	petH := handler.NewPetHandler(petSvc, logger)
	postH := handler.NewPostHandler(postSvc, feedSvc, logger)
	meetupH := handler.NewMeetupHandler(meetupSvc, userRepo, logger)
	chatH := handler.NewChatHandler(chatSvc, logger)
	feedH := handler.NewFeedHandler(feedSvc, logger)
	auditH := handler.NewAuditHandler(auditSvc, logger)
	var mediaH *handler.MediaHandler
	if mediaSvc != nil {
		mediaH = handler.NewMediaHandler(mediaSvc, logger)
	}

	engine := router.New(&router.Deps{
		Cfg:       cfg,
		Logger:    logger,
		Repo:      &router.RepositoryBundle{User: userRepo, Pet: petRepo, Post: postRepo, Meetup: meetupRepo, Chat: chatRepo, Audit: auditRepo},
		Service:   &router.ServiceBundle{User: userSvc, Pet: petSvc, Post: postSvc, Meetup: meetupSvc, Chat: chatSvc, Audit: auditSvc, Feed: feedSvc, Media: mediaSvc},
		Handler:   &router.HandlerBundle{User: userH, Pet: petH, Post: postH, Meetup: meetupH, Chat: chatH, Audit: auditH, Feed: feedH, Media: mediaH},
		Redis:     rdb,
		JWTSecret: cfg.JWTSecret,
	})

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: engine,
	}
	go func() {
		logger.Info(fmt.Sprintf("server listening on :%s", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("server stopping")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func logServerStart(cfg *config.Config) string {
	return fmt.Sprintf("server starting env=%s port=%s", cfg.AppEnv, cfg.HTTPPort)
}

func redisConnected(cfg *config.Config) string {
	return fmt.Sprintf("redis connected addr=%s", cfg.RedisAddr)
}
