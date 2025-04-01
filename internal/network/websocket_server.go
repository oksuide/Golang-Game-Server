package network

import (
	"log"
	"net/http"
	"sync/atomic"

	"gameCore/internal/game"

	"github.com/gorilla/websocket"
)

// WebSocket-сервер
type WebSocketServer struct {
	Game *game.Game
}

func NewWebSocketServer(gameInstance *game.Game) *WebSocketServer {
	return &WebSocketServer{
		Game: gameInstance,
	}
}

// Обработчик подключения
func (s *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка WebSocket:", err)
		return
	}

	// Генерируем уникальный ID игрока (автоинкремент)
	playerID := uint(atomic.AddUint32(&playerCounter, 1))

	// Добавляем игрока в игру
	s.Game.AddPlayer(playerID, conn)
	log.Printf("Игрок %d подключился\n", playerID)

	// Запускаем обработку сообщений от игрока
	go s.listenToPlayer(playerID, conn)
}

// Счетчик игроков (автоинкремент)
var playerCounter uint32 = 0

// Прослушивание сообщений от игрока
func (s *WebSocketServer) listenToPlayer(playerID uint, conn *websocket.Conn) {
	defer func() {
		s.Game.RemovePlayer(playerID) // Удаляем игрока при отключении
		conn.Close()
		log.Printf("Игрок %d отключился\n", playerID)
	}()

	for {
		var input struct {
			Up, Down, Left, Right bool
			Angle                 float64
		}

		// Читаем сообщение от игрока
		err := conn.ReadJSON(&input)
		if err != nil {
			log.Printf("Ошибка чтения от игрока %d: %v", playerID, err)
			return
		}

		// Обновляем игрока в игре
		s.Game.UpdatePlayer(playerID, input)
	}
}

// BroadcastGameState рассылает состояние игры всем игрокам
func (s *WebSocketServer) BroadcastGameState() {
	s.Game.RLock() // Блокируем чтение, чтобы избежать гонок данных
	defer s.Game.RUnlock()

	state := make(map[uint]game.PlayerState)

	// Собираем информацию о всех игроках
	for id, player := range s.Game.Players {
		state[id] = game.PlayerState{
			X:     player.X,
			Y:     player.Y,
			Angle: player.Angle,
		}
	}

	// Отправляем состояние всем подключённым игрокам
	for _, player := range s.Game.Players {
		if player.Conn != nil {
			err := player.Conn.WriteJSON(state)
			if err != nil {
				log.Printf("Ошибка отправки данных игроку %d: %v", player.ID, err)
				player.Conn.Close()
			}
		}
	}
}

// Запуск WebSocket-сервера
func (s *WebSocketServer) Start(address string) {
	http.HandleFunc("/ws", s.HandleConnections)

	log.Println("Запуск WebSocket-сервера на", address)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("Ошибка запуска WebSocket-сервера:", err)
	}
}
