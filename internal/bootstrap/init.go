package bootstrap

import (
	"log"

	"gameCore/config"
	"gameCore/internal/game"
	"gameCore/internal/network"
	"gameCore/internal/storage"
)

// Init запускает все сервисы и возвращает игровые объекты
func Init() (*game.Game, *network.WebSocketServer) {
	// Загружаем конфиг
	cfg, err := config.LoadConfig("/home/oksuide/GoProjects/gameCore/config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Подключаемся к БД
	if err := storage.Connect(cfg.Database); err != nil {
		log.Fatal("Ошибка БД:", err)
	}
	if err := storage.InitTables(); err != nil {
		log.Fatal("Ошибка миграции БД:", err)
	}

	// Инициализируем Redis
	if err := storage.InitRedis(cfg.Redis); err != nil {
		log.Fatal("Ошибка Redis:", err)
	}

	// Создаем игровой движок
	gameInstance := game.NewGame()

	// Создаем WebSocket-сервер
	wsServer := network.NewWebSocketServer(gameInstance)

	return gameInstance, wsServer
}
