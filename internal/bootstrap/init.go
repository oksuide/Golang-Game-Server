package bootstrap

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"gameCore/internal/auth"
	"gameCore/internal/config"
	"gameCore/internal/game"
	"gameCore/internal/middleware"
	"gameCore/internal/network"
	"gameCore/internal/repository"
	"gameCore/internal/storage"
	"gameCore/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Init() (*game.Game, *network.WebSocketServer, *gin.Engine) {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		os.Exit(1)
	}

	log, err := utils.SetupLogger(cfg.App.Env)
	if err != nil {
		log.Error("Logger setup failed", "error", err)
		os.Exit(1)
	}

	log.Info("Starting application initialization", slog.String("env", cfg.App.Env))

	// Database initialization
	if err := storage.Connect(cfg.Database); err != nil {
		log.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	if err := storage.InitTables(); err != nil {
		log.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	// Repository initialization
	userRepo := repository.NewUserRepo(storage.DB)
	// leaderboardRepo := repository.NewLeaderboardRepo(storage.DB)
	// playerRepo := repository.NewPlayerRepo(storage.DB)

	// Redis initialization
	if err := storage.InitRedis(cfg.Redis); err != nil {
		log.Error("Redis initialization failed", "error", err)
		os.Exit(1)
	}

	// Auth handler setup
	authHandler := auth.NewAuthHandler(userRepo, cfg)
	// WebSocket server

	// Game core initialization
	gameInstance := game.NewGame(
	// game.WithPlayerRepo(playerRepo),
	// game.WithLeaderboardRepo(leaderboardRepo),
	)

	wsServer := network.NewWebSocketServer(gameInstance, cfg.WebSocket)

	// Router setup
	router := gin.Default()
	setupRoutes(router, authHandler, wsServer, cfg.JWT)

	log.Info("Application initialization completed")
	return gameInstance, wsServer, router
}

func setupRoutes(router *gin.Engine, authHandler *auth.AuthHandler, wsServer *network.WebSocketServer, jwtSecret config.JWTConfig) {
	// Настройка CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // URL вашего фронтенда
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	api := router.Group("/api")
	{
		api.POST("/register", authHandler.RegisterHandler)
		api.POST("/login", authHandler.LoginHandler)
	}

	authorized := api.Group("")
	authorized.Use(middleware.AuthMiddleware(jwtSecret.SecretKey))
	{
		// Регистрируем WebSocket endpoint в защищенной группе
		authorized.GET("/ws", func(c *gin.Context) {
			userID, exists := c.Get("userID")
			if !exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}

			wsServer.HandleWS(c.Writer, c.Request, userID.(uint))
		})
	}
}
