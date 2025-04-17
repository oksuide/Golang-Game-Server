package network

import (
	"gameCore/internal/config"
	"gameCore/internal/game"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	Game     *game.Game
	Config   config.WebSocketConfig
	upgrader websocket.Upgrader
}

func NewWebSocketServer(gameInstance *game.Game, wsConfig config.WebSocketConfig) *WebSocketServer {
	if wsConfig.ReadBufferSize == 0 {
		wsConfig.ReadBufferSize = 4096
	}
	if wsConfig.WriteBufferSize == 0 {
		wsConfig.WriteBufferSize = 4096
	}
	if wsConfig.PongTimeout == 0 {
		wsConfig.PongTimeout = 60 * time.Second
	}

	return &WebSocketServer{
		Game:   gameInstance,
		Config: wsConfig,
		upgrader: websocket.Upgrader{
			ReadBufferSize:   wsConfig.ReadBufferSize,
			WriteBufferSize:  wsConfig.WriteBufferSize,
			CheckOrigin:      func(r *http.Request) bool { return true },
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

func (s *WebSocketServer) RegisterRoutes(router *gin.Engine) {
	router.GET("/ws", func(c *gin.Context) {
		// Извлекаем userID из контекста, установленного middleware
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Вызов обработчика WebSocket
		s.HandleWS(c.Writer, c.Request, userID.(uint))
	})
}

func (s *WebSocketServer) HandleWS(w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade error", http.StatusInternalServerError)
		return
	}

	// Закрытие соединения при выходе
	defer conn.Close()

	// Используем userID из аутентификации
	playerID := userID

	err = s.Game.AddPlayer(playerID, conn)
	if err != nil {
		log.Println("Ошибка добавления игрока:", err)
		return
	}

	// Отправка ID игрока
	conn.WriteJSON(map[string]interface{}{
		"yourId": playerID,
	})

	// Обработка входящих сообщений
	go func() {
		defer s.Game.RemovePlayer(playerID)

		for {
			var input game.PlayerInputData
			if err := conn.ReadJSON(&input); err != nil {
				log.Printf("Ошибка чтения от игрока %d: %v", playerID, err)
				return
			}
			s.Game.Inputs <- game.PlayerInput{
				ID:    playerID,
				Input: input,
			}
		}
	}()
}
