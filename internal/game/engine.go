package game

import (
	"errors"
	"log"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Константы для настройки игры
const (
	GameTick          = 16 * time.Millisecond // ~60 FPS
	MaxInputQueue     = 1000                  // Увеличим буфер канала ввода
	BasePlayerSpeed   = 10.0
	BasePlayerHealth  = 100.0
	CollisionDistance = 10.0
	BulletLifetime    = 5 * time.Second
)

// PlayerInput — структура для событий от игроков
type PlayerInputData struct {
	Up, Down, Left, Right bool
	Angle                 float64
	Shoot                 bool
}

type PlayerInput struct {
	ID          uint
	Input       PlayerInputData
	UpgradeStat string
}

// Game — игровое ядро с дополнительными полями для управления игрой
type Game struct {
	Players map[uint]*Player
	Mutex   sync.RWMutex
	Inputs  chan PlayerInput
	Bullets []*Bullet
	Running bool          // Флаг работы игрового цикла
	Done    chan struct{} // Канал для остановки игры
}

// NewGame создает новую игру с инициализированными полями
func NewGame() *Game {
	return &Game{
		Players: make(map[uint]*Player),
		Inputs:  make(chan PlayerInput, MaxInputQueue),
		Done:    make(chan struct{}),
		Running: false,
	}
}

// AddPlayer добавляет нового игрока с базовыми характеристиками
func (g *Game) AddPlayer(id uint, conn *websocket.Conn) error {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	// Проверяем, не существует ли уже игрок с таким ID
	if _, exists := g.Players[id]; exists {
		return errors.New("игрок с таким ID уже существует")
	}

	g.Players[id] = &Player{
		ID:          id,
		X:           0,
		Y:           0,
		Conn:        conn,
		Level:       1,
		XP:          0,
		SkillPoints: 0,
		Stats: map[string]float64{
			"health":       BasePlayerHealth,
			"damage":       10,
			"speed":        BasePlayerSpeed,
			"fire_rate":    1.0,
			"body_damage":  10,
			"bullet_speed": 10,
			"reload_speed": 3,
		},
		lastShot: time.Now(),
	}
	log.Printf("Добавлен игрок %d", id)
	return nil
}

// RemovePlayer удаляет игрока из игры
func (g *Game) RemovePlayer(id uint) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if player, exists := g.Players[id]; exists {
		if player.Conn != nil {
			player.Conn.Close()
		}
		delete(g.Players, id)
		log.Printf("Игрок %d удален", id)
	}
}

// Start запускает игровой цикл
func (g *Game) Start() {
	if g.Running {
		return
	}

	g.Running = true
	ticker := time.NewTicker(GameTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.updateGameState()
		case input := <-g.Inputs:
			g.processInput(input)
		case <-g.Done:
			g.Running = false
			return
		}
	}
}

// Stop останавливает игровой цикл
func (g *Game) Stop() {
	if g.Running {
		close(g.Done)
	}
}

// updateGameState обновляет состояние игры
func (g *Game) updateGameState() {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	// Обновляем игроков
	for _, player := range g.Players {
		player.LevelUp()
	}

	// Обновляем пули
	g.updateBullets()

	// Отправляем обновления игрокам
	g.sendUpdates()
}

// sendUpdates отправляет обновления состояния всем игрокам
func (g *Game) sendUpdates() {
	// Создаем копию состояния для отправки
	playersCopy := make(map[uint]*Player, len(g.Players))
	for id, player := range g.Players {
		playersCopy[id] = player
	}

	for _, player := range g.Players {
		if player.Conn == nil {
			continue
		}

		// Отправляем только необходимые данные, а не всех игроков
		gameState := struct {
			Players map[uint]*Player `json:"players"`
			Bullets []*Bullet        `json:"bullets"`
		}{
			Players: playersCopy,
			Bullets: g.Bullets,
		}

		if err := player.Conn.WriteJSON(gameState); err != nil {
			log.Printf("Ошибка отправки данных игроку %d: %v", player.ID, err)
			player.Conn.Close()
			player.Conn = nil
		}
	}
}

// processInput обрабатывает ввод игрока
func (g *Game) processInput(input PlayerInput) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	player, exists := g.Players[input.ID]
	if !exists {
		return
	}

	g.handleMovement(player, input.Input)
	g.handleRotation(player, input.Input)
	g.handleShooting(player, input.Input)
	g.handleUpgrades(player, input.UpgradeStat)
}

// handleMovement обрабатывает движение игрока
func (g *Game) handleMovement(player *Player, input struct {
	Up, Down, Left, Right bool
	Angle                 float64
	Shoot                 bool
}) {
	// Нормализуем движение по диагонали
	var moveX, moveY float64
	speed := player.Stats["speed"]

	if input.Up && !input.Down {
		moveY = -speed
	} else if input.Down && !input.Up {
		moveY = speed
	}

	if input.Left && !input.Right {
		moveX = -speed
	} else if input.Right && !input.Left {
		moveX = speed
	}

	// Если движение по диагонали, нормализуем вектор
	if moveX != 0 && moveY != 0 {
		length := math.Sqrt(moveX*moveX + moveY*moveY)
		moveX = moveX / length * speed
		moveY = moveY / length * speed
	}

	player.X += moveX
	player.Y += moveY

	// Проверяем границы игрового поля (можно добавить константы для размеров поля)
	const worldSize = 1000.0
	player.X = math.Max(-worldSize, math.Min(worldSize, player.X))
	player.Y = math.Max(-worldSize, math.Min(worldSize, player.Y))
}

// handleRotation обрабатывает поворот игрока
func (g *Game) handleRotation(player *Player, input struct {
	Up, Down, Left, Right bool
	Angle                 float64
	Shoot                 bool
}) {
	player.Angle = input.Angle
}
