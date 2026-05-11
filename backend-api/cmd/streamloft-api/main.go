package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"streamloft-api/internal/config"
	"streamloft-api/internal/crypto"
	"streamloft-api/internal/database"
	"streamloft-api/internal/handler"
	"streamloft-api/internal/logger"
	"streamloft-api/internal/middleware"
	"streamloft-api/internal/repository"
	"streamloft-api/internal/service"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.LogLevel)
	l := logger.Get()
	l.Info().Msg("StreamLoft API starting")

	enc, err := crypto.NewEncryptor(cfg.EncryptionKeyPath)
	if err != nil {
		l.Fatal().Err(err).Msg("encryption init failed")
	}

	db, err := database.New(context.Background(), cfg.DatabaseURL())
	if err != nil {
		l.Fatal().Err(err).Msg("database connect failed")
	}
	defer db.Close()

	if err := db.RunMigrations(context.Background()); err != nil {
		l.Fatal().Err(err).Msg("migrations failed")
	}

	pool := db.Pool()
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	machineRepo := repository.NewMachineRepository(pool)
	broadcastRepo := repository.NewBroadcastRepository(pool)

	jwtSvc := service.NewJWTService(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo, sessionRepo, machineRepo, enc, jwtSvc, cfg)
	authHdlr := handler.NewAuthHandler(authSvc, cfg)

	userSvc := service.NewUserService(userRepo)
	userHdlr := handler.NewUserHandler(userSvc, cfg)

	workerSvc := service.NewWorkerService(userRepo, enc, cfg.RTMPURL)
	streamSvc := service.NewStreamService(userRepo, broadcastRepo, workerSvc)
	streamHdlr := handler.NewStreamHandler(streamSvc, userSvc, workerSvc, cfg)

	destSvc := service.NewDestinationService(userRepo, workerSvc, streamSvc.IsUserLive, enc)
	destHdlr := handler.NewDestinationHandler(destSvc)

	broadcastSvc := service.NewBroadcastService(broadcastRepo)
	broadcastHdlr := handler.NewBroadcastHandler(broadcastSvc)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", authHdlr.Login)
		authGroup.POST("/refresh", authHdlr.Refresh)
		authGroup.POST("/logout", authHdlr.Logout)
	}

	streamGroup := r.Group("/stream")
	{
		streamGroup.POST("/start", streamHdlr.Start)
		streamGroup.POST("/stop", streamHdlr.Stop)
	}

	userGroup := r.Group("/user")
	userGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		userGroup.GET("", userHdlr.GetUser)
		userGroup.PUT("/bitrate", streamHdlr.UpdateBitrate)
		userGroup.GET("/stream/status", streamHdlr.Status)
	}

	destGroup := r.Group("/destinations")
	destGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		destGroup.GET("", destHdlr.List)
		destGroup.PUT("/:id", destHdlr.Update)
	}

	broadcastGroup := r.Group("/broadcasts")
	broadcastGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		broadcastGroup.GET("", broadcastHdlr.List)
	}

	addr := ":" + cfg.APIPort
	l.Info().Str("addr", addr).Msg("server starting")
	r.Run(addr)
}