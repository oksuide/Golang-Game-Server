package network

import (
	"gameCore/internal/config"
	"gameCore/internal/game"
	"log"
	"net/http"
	"time"

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

func (s *WebSocketServer) HandleWS(w http.ResponseWriter, r *http.Request, userID uint) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Добавляем игрока с аутентифицированным ID
	err = s.Game.AddPlayer(userID, conn) // Используем userID из middleware
	if err != nil {
		log.Printf("Add player error: %v", err)
		conn.WriteJSON(map[string]interface{}{
			"error": "Failed to join game",
		})
		return
	}

	// Уведомление об успешном подключении
	conn.WriteJSON(map[string]interface{}{
		"yourId": userID,
		"status": "connected",
	})

	// Обработчик входящих сообщений
	go s.handleMessages(conn, userID)
}

func (s *WebSocketServer) handleMessages(conn *websocket.Conn, userID uint) {
	defer s.Game.RemovePlayer(userID)

	for {
		var input game.PlayerInputData
		if err := conn.ReadJSON(&input); err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("Player %d disconnected: %v", userID, err)
			}
			return
		}

		s.Game.Inputs <- game.PlayerInput{
			ID:    userID,
			Input: input,
		}
	}
}
