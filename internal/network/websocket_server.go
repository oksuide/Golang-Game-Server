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
	Game       *game.Game
	httpServer *http.Server
	Config     config.WebSocketConfig
	upgrader   websocket.Upgrader
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

func (s *WebSocketServer) StartServer() {
	http.HandleFunc("/ws", s.handleWS)

	s.httpServer = &http.Server{
		Addr:    "localhost:8080",
		Handler: nil,
	}

	go func() {
		log.Printf("🌐 WebSocket сервер слушает на localhost:8080")
		if err := s.httpServer.ListenAndServe(); err != nil {
			log.Fatalf("WebSocket сервер упал: %v", err)
		}
	}()
}

func (s *WebSocketServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade error", http.StatusInternalServerError)
		return
	}

	// Простой ID (в реальности должен быть от авторизации)
	playerID := uint(time.Now().UnixNano() % 1000000)

	err = s.Game.AddPlayer(playerID, conn)
	if err != nil {
		log.Println("Ошибка добавления игрока:", err)
		conn.Close()
		return
	}

	conn.WriteJSON(map[string]interface{}{
		"yourId": playerID,
	})

	go func() {
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
