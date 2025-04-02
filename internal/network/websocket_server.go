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

type WebSocketServer struct {
	Game       *game.Game
	httpServer *http.Server
	Config     config.WebSocketConfig
	upgrader   websocket.Upgrader
}

// NewWebSocketServer создает новый WebSocket сервер
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
			ReadBufferSize:  wsConfig.ReadBufferSize,
			WriteBufferSize: wsConfig.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // В production заменить на проверку origin
			},
		},
	}
}

var playerCounter uint32 = 0

func (s *WebSocketServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	playerID := uint(atomic.AddUint32(&playerCounter, 1))

	// Обрабатываем отключение даже если не дошли до handlePlayerMessages
	defer s.handleDisconnection(playerID, conn)

	conn.SetReadLimit(int64(s.Config.MaxMessageSize))
	conn.SetReadDeadline(time.Now().Add(s.Config.PongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(s.Config.PongTimeout))
		return nil
	})

	s.Game.AddPlayer(playerID, conn)

	log.Printf("Player %d connected", playerID)
	s.handlePlayerMessages(playerID, conn)
}

func (s *WebSocketServer) handlePlayerMessages(playerID uint, conn *websocket.Conn) {
	defer s.handleDisconnection(playerID, conn)

	for {
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
				log.Printf("Player %d read error: %v", playerID, err)
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

func (s *WebSocketServer) handleDisconnection(playerID uint, conn *websocket.Conn) {
	s.Game.RemovePlayer(playerID)
	conn.Close()
	log.Printf("Player %d disconnected", playerID)
}

func (s *WebSocketServer) Start(address string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleConnection)

	s.httpServer = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("Starting WebSocket server on %s", address)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("WebSocket server failed: %v", err)
	}
}

func (s *WebSocketServer) Shutdown(ctx context.Context) error {
	log.Println("Shutting down WebSocket server...")
	return s.httpServer.Shutdown(ctx)
}
