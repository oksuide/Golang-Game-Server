package game

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Размеры карты
const (
	MapWidth     = 2000
	MapHeight    = 2000
	TickRate     = 16 * time.Millisecond // ~60 тиков в секунду
	MaxSpeed     = 200.0                 // Пикселей в секунду
	Acceleration = 5.0                   // Скорость набора/снижения
)

// Вспомогательная функция для интерполяции
func lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

type InputMessage struct {
	Up, Down, Left, Right bool
	Angle                 float64
}

// Игрок
type Player struct {
	ID          uint
	X, Y        float64 // Позиция
	Vx, Vy      float64 // Скорость по осям
	Angle       float64 // Направление
	Conn        *websocket.Conn
	lastUpdated time.Time
}

// Игровой мир (ECS)
type Game struct {
	Players        map[uint]*Player
	MovementSystem *MovementSystem
	NetworkSystem  *NetworkSystem
	mu             sync.RWMutex
}

// Создаем новую игру
func NewGame() *Game {
	return &Game{
		Players:        make(map[uint]*Player),
		MovementSystem: &MovementSystem{},
		NetworkSystem:  &NetworkSystem{},
	}
}

// Добавляем игрока
func (g *Game) AddPlayer(id uint, conn *websocket.Conn) {
	g.mu.Lock()
	defer g.mu.Unlock()

	time.Now().UnixNano()
	x := rand.Float64() * MapWidth
	y := rand.Float64() * MapHeight

	g.Players[id] = &Player{
		ID:          id,
		X:           x,
		Y:           y,
		Vx:          0,
		Vy:          0,
		Conn:        conn,
		lastUpdated: time.Now(),
	}
}

// Основная игровая петля
func (g *Game) GameLoop() {
	ticker := time.NewTicker(TickRate)
	defer ticker.Stop()

	for range ticker.C {
		g.Update()
	}
}

// Обновление игры (ECS)
func (g *Game) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	delta := TickRate.Seconds()

	// Получаем данные от клиентов
	g.NetworkSystem.Update(g.Players)

	// Обновляем физику
	g.MovementSystem.Update(g.Players, delta)
}

// Обновляет игрока, скрывая мьютекс внутри
func (g *Game) UpdatePlayer(playerID uint, input InputMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()

	player, exists := g.Players[playerID]
	if exists {
		g.NetworkSystem.UpdatePlayer(player, input)
	}
}

func (g *Game) Lock() {
	g.mu.Lock()
}

func (g *Game) Unlock() {
	g.mu.Unlock()
}

func (g *Game) RLock() {
	g.mu.RLock()
}

func (g *Game) RUnlock() {
	g.mu.RUnlock()
}

// ------------------ SYSTEMS ------------------

// Система движения
type MovementSystem struct{}

func (s *MovementSystem) Update(players map[uint]*Player, delta float64) {
	for _, p := range players {
		if p.Conn == nil {
			continue
		}

		// Плавное изменение скорости
		p.X += p.Vx * delta
		p.Y += p.Vy * delta

		// Ограничение по карте
		if p.X < 0 {
			p.X = 0
		} else if p.X > MapWidth {
			p.X = MapWidth
		}
		if p.Y < 0 {
			p.Y = 0
		} else if p.Y > MapHeight {
			p.Y = MapHeight
		}
	}
}

// Система получения входных данных от клиентов
type NetworkSystem struct{}

func (s *NetworkSystem) Update(players map[uint]*Player) {
	for _, p := range players {
		if p.Conn == nil {
			continue
		}

		// Получаем входные данные
		var input struct {
			Up, Down, Left, Right bool
			Angle                 float64
		}

		err := p.Conn.ReadJSON(&input)
		if err != nil {
			log.Println("Ошибка чтения данных:", err)
			continue
		}

		// Обновляем игрока через отдельную функцию
		s.UpdatePlayer(p, input)
	}
}

// Обновление конкретного игрока
func (s *NetworkSystem) UpdatePlayer(player *Player, input struct {
	Up, Down, Left, Right bool
	Angle                 float64
}) {
	// Обновляем направление
	player.Angle = input.Angle

	// Вычисляем целевые значения скорости
	targetVx, targetVy := 0.0, 0.0
	if input.Up {
		targetVy = -MaxSpeed
	}
	if input.Down {
		targetVy = MaxSpeed
	}
	if input.Left {
		targetVx = -MaxSpeed
	}
	if input.Right {
		targetVx = MaxSpeed
	}

	// Плавное изменение скорости
	player.Vx = lerp(player.Vx, targetVx, Acceleration*TickRate.Seconds())
	player.Vy = lerp(player.Vy, targetVy, Acceleration*TickRate.Seconds())
}

// Удаление игрока из игры
func (g *Game) RemovePlayer(id uint) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Players, id)
}

// Получение состояния игры
func (g *Game) GetState() map[uint]PlayerState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	state := make(map[uint]PlayerState)
	for id, p := range g.Players {
		state[id] = PlayerState{
			X:     p.X,
			Y:     p.Y,
			Vx:    p.Vx,
			Vy:    p.Vy,
			Angle: p.Angle,
		}
	}
	return state
}

// Структура для передачи состояния
type PlayerState struct {
	X, Y, Vx, Vy, Angle float64
}
