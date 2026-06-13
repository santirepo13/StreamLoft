package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	workerSvc := service.NewWorkerService(userRepo, enc, cfg.SRSURL)
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
		userGroup.GET("/stream/events", streamHdlr.Events)
	}

	destGroup := r.Group("/destinations")
	destGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		destGroup.GET("", destHdlr.List)
		destGroup.PUT("/:id", destHdlr.Update)
		destGroup.PUT("/toggle/:id", destHdlr.Toggle)
	}

	broadcastGroup := r.Group("/broadcasts")
	broadcastGroup.Use(middleware.AuthMiddleware(jwtSvc))
	{
		broadcastGroup.GET("", broadcastHdlr.List)
	}

	addr := ":" + cfg.APIPort
	l.Info().Str("addr", addr).Msg("server starting")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	server := &http.Server{Addr: addr, Handler: r}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("server failed")
		}
	}()

	l.Info().Msg("server started, waiting for shutdown signal")

	// Wait for interrupt signal
	<-sigChan
	l.Info().Msg("received shutdown signal, gracefully shutting down...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop all workers gracefully
	if err := workerSvc.StopAllWorkers(); err != nil {
		l.Error().Err(err).Msg("failed to stop all workers")
	}

	// Shutdown the HTTP server
	if err := server.Shutdown(ctx); err != nil {
		l.Error().Err(err).Msg("server shutdown failed")
	}

	l.Info().Msg("server shutdown completed")
}