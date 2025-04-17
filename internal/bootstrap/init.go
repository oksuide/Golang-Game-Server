package bootstrap

import (
	"log/slog"
	"net/http"
	"os"

	"gameCore/internal/auth"
	"gameCore/internal/config"
	"gameCore/internal/game"
	"gameCore/internal/network"
	"gameCore/internal/repository"
	"gameCore/internal/storage"
	"gameCore/internal/utils"

	"github.com/gin-gonic/gin"
)

func Init() (*game.Game, *network.WebSocketServer, *gin.Engine) {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		os.Exit(1)
	}

	log, err := utils.SetupLogger(cfg.App.Env)
	if err != nil {
		log.Error("Logger setup failed", err)
		os.Exit(1)
	}

	log.Info("Starting application initialization", slog.String("env", cfg.App.Env))

	// Database initialization
	if err := storage.Connect(cfg.Database); err != nil {
		log.Error("Database connection failed", err)
		os.Exit(1)
	}

	if err := storage.InitTables(); err != nil {
		log.Error("Database migration failed", err)
		os.Exit(1)
	}

	// Repository initialization
	userRepo := repository.NewUserRepo(storage.DB)
	// leaderboardRepo := repository.NewLeaderboardRepo(storage.DB)
	// playerRepo := repository.NewPlayerRepo(storage.DB)

	// Redis initialization
	if err := storage.InitRedis(cfg.Redis); err != nil {
		log.Error("Redis initialization failed", err)
		os.Exit(1)
	}

	// Auth handler setup
	authHandler := auth.NewAuthHandler(userRepo, cfg)

	// Router setup
	router := gin.Default()
	setupRoutes(router, authHandler)

	// Game core initialization
	gameInstance := game.NewGame(
	// game.WithPlayerRepo(playerRepo),
	// game.WithLeaderboardRepo(leaderboardRepo),
	)

	// WebSocket server
	wsServer := network.NewWebSocketServer(gameInstance, cfg.WebSocket)

	log.Info("Application initialization completed")
	return gameInstance, wsServer, router
}

func setupRoutes(router *gin.Engine, authHandler *auth.AuthHandler) {
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)
	}
}
