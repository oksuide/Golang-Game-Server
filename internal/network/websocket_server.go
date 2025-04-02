package network

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"gameCore/config"
	"gameCore/internal/game"

	"github.com/gorilla/websocket"
)

// WebSocketServer представляет сервер WebSocket для обработки соединений игроков
type WebSocketServer struct {
	Game       *game.Game
	httpServer *http.Server
	Config     config.WebSocketConfig
	upgrader   websocket.Upgrader
}

// NewWebSocketServer создает новый WebSocket сервер с настроенными параметрами
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

var playerCounter uint32 = 0

// handleConnection обрабатывает новое WebSocket соединение
func (s *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка обновления WebSocket: %v", err)
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("Аномальное закрытие соединения: %v", err)
		}
		return
	}

	playerID := uint(atomic.AddUint32(&playerCounter, 1))

	defer s.handleDisconnection(playerID, conn)

	conn.SetReadLimit(int64(s.Config.MaxMessageSize))
	conn.SetReadDeadline(time.Now().Add(s.Config.PongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(s.Config.PongTimeout))
		return nil
	})

	if err := s.Game.AddPlayer(playerID, conn); err != nil {
		log.Printf("Ошибка добавления игрока: %v", err)
		conn.Close()
		return
	}

	log.Printf("Игрок %d подключился", playerID)
	s.handlePlayerMessages(playerID, conn)
}

// handlePlayerMessages обрабатывает сообщения от игрока
func (s *WebSocketServer) handlePlayerMessages(playerID uint, conn *websocket.Conn) {
	defer s.handleDisconnection(playerID, conn)

	for {
		select {
		case <-context.Background().Done():
			return
		default:
			var input struct {
				Up          bool    `json:"up"`
				Down        bool    `json:"down"`
				Left        bool    `json:"left"`
				Right       bool    `json:"right"`
				Angle       float64 `json:"angle"`
				Shoot       bool    `json:"shoot"`
				UpgradeStat string  `json:"upgradeStat,omitempty"`
			}

			if err := conn.ReadJSON(&input); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Ошибка чтения от игрока %d: %v", playerID, err)
				}
				return
			}

			s.Game.Inputs <- game.PlayerInput{
				ID: playerID,
				Input: game.PlayerInputData{
					Up:    input.Up,
					Down:  input.Down,
					Left:  input.Left,
					Right: input.Right,
					Angle: input.Angle,
					Shoot: input.Shoot,
				},
				UpgradeStat: input.UpgradeStat,
			}
		}
	}
}

// handleDisconnection обрабатывает отключение игрока
func (s *WebSocketServer) handleDisconnection(playerID uint, conn *websocket.Conn) {
	s.Game.RemovePlayer(playerID)
	conn.Close()
	log.Printf("Игрок %d отключился", playerID)
}

// Start запускает WebSocket сервер
func (s *WebSocketServer) Start(address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleConnection)

	s.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Printf("Запуск WebSocket сервера на %s", address)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Ошибка запуска WebSocket сервера: %v", err)
	}
}

// Shutdown останавливает WebSocket сервер
func (s *WebSocketServer) Shutdown(ctx context.Context) error {
	log.Println("Остановка WebSocket сервера...")
	return s.httpServer.Shutdown(ctx)
}
