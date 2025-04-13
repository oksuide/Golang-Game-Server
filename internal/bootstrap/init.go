package bootstrap

import (
	"log/slog"
	"os"

	"gameCore/internal/config"
	"gameCore/internal/game"
	"gameCore/internal/network"
	"gameCore/internal/storage"
	"gameCore/internal/utils"
)

func Init() (*game.Game, *network.WebSocketServer) {
	// Загружаем конфиг
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		os.Exit(1)
	}

	// Настройка логгера
	log, err := utils.SetupLogger(cfg.App.Env)
	if err != nil {
		log.Error("Logger setup failed", err)
		os.Exit(1)
	}

	// Логирование начала инициализации
	log.Info("Starting application initialization", slog.String("env", cfg.App.Env))

	log.Debug("Debug messages are enabled")
	// Подключаемся к БД
	if err := storage.Connect(cfg.Database); err != nil {
		log.Error("Database connection failed", err)
		os.Exit(1)
	}

	// Инициализируем таблицы
	if err := storage.InitTables(); err != nil {
		log.Error("Database migration failed", err)
		os.Exit(1)
	}

	// Инициализируем Redis
	if err := storage.InitRedis(cfg.Redis); err != nil {
		log.Error("Redis initialization failed", err)
		os.Exit(1)
	}

	// router := routes.SetupRouter()

	// Создаем игровой движок
	log.Info("Initializing game engine")
	gameInstance := game.NewGame()

	// Создаем WebSocket-сервер
	log.Info("Initializing WebSocket server")
	wsServer := network.NewWebSocketServer(gameInstance, cfg.WebSocket)

	log.Info("Application initialization completed successfully")
	return gameInstance, wsServer
}
