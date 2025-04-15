package game

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// TODO:
// 3. Добавить механизм для обработки столкновений (чуть позже) t2
// 6. Добавить механизм для обработки улучшений t1
// 7. Добавить механизм для обработки специлизаций t2
// 8. Добавить объекты на карте для фарма опыта t3
// 9. Сделать более плавное передвижение t4
// 10. Сделать фиксированное положение камеры t4
// 11. Сделать более плавную анимацию t4
// 12. Сделать более плавную анимацию выстрелов t4

const (
	GameTick          = 16 * time.Millisecond // ~60 FPS
	MaxInputQueue     = 1000                  // Буфер канала ввода
	BasePlayerSpeed   = 10.0
	BasePlayerHealth  = 100.0
	CollisionDistance = 10.0
	PlayerRadius      = 10.0 // Радиус игрока
	BulletRadius      = 3.0  // Радиус пули
	ObjectRadius      = 15.0
	RespawnTime       = 5 * time.Second
	MinX              = 0
	MaxX              = 1880
	MinY              = 0
	MaxY              = 1040
	// BulletLifetime    = 5 * time.Second
)

type Game struct {
	Players      map[uint]*Player
	Objects      []*Object
	Mutex        sync.RWMutex
	Inputs       chan PlayerInput
	Bullets      []*Bullet
	RespawnDelay time.Duration
	MaxObjects   int
	Running      bool          // Флаг работы игрового цикла
	Done         chan struct{} // Канал для остановки игры
}

type Player struct {
	ID    uint
	X, Y  float64
	Angle float64         `json:"angle"`
	Conn  *websocket.Conn `json:"-"`

	Level            int
	XP               int
	NewLvlExp        int
	SkillPoints      int `json:"skill_points"`
	Stats            map[string]float64
	Alive            bool
	FailedBroadcasts int
	RespawnTimer     *time.Timer `json:"-"`
	lastShot         time.Time   // Время последнего выстрела для контроля скорострельности
}

type Bullet struct {
	ID        uint
	OwnerID   uint
	X, Y      float64
	Angle     float64
	Speed     float64
	Damage    float64
	Active    bool
	CreatedAt time.Time // Добавляем время создания для возможного времени жизни пули
}

type bulletState struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
}

type playerState struct {
	ID               uint               `json:"id"`
	X                float64            `json:"x"`
	Y                float64            `json:"y"`
	Angle            float64            `json:"angle"`
	Level            int                `json:"level"`
	NewLvlExp        int                `json:"new_lvl_epx"`
	SkillPoints      int                `json:"skill_points"`
	FailedBroadcasts int                `json:"failed_broadcasts"`
	Stats            map[string]float64 `json:"stats"`
}

type objectState struct {
	ID uint `json:"id"`
	// Type   string  `json:"type"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Health int     `json:"health"`
}

type PlayerInputData struct {
	Up, Down, Left, Right bool
	Angle                 float64 `json:"angle"`
	Shoot                 bool
	UpgradeStat           string `json:"stat"`
}

type PlayerInput struct {
	ID    uint
	Input PlayerInputData
}

func NewGame() *Game {
	game := &Game{
		Players:      make(map[uint]*Player),
		Inputs:       make(chan PlayerInput, MaxInputQueue),
		Done:         make(chan struct{}),
		Running:      false,
		Objects:      make([]*Object, 0),
		MaxObjects:   30,              // default object count
		RespawnDelay: 1 * time.Minute, // default respawn time
	}
	game.InitObjectSystem(game.MaxObjects, game.RespawnDelay)
	return game
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
		ID:               id,
		X:                float64(rand.Intn(MaxX) + 60),
		Y:                float64(rand.Intn(MaxY) + 40),
		Conn:             conn,
		Level:            1,
		XP:               0,
		NewLvlExp:        100,
		SkillPoints:      0,
		FailedBroadcasts: 0,
		Alive:            true,
		Stats: map[string]float64{
			"health":     BasePlayerHealth,
			"max_health": BasePlayerHealth,
			"damage":     10,
			"speed":      40,
			"fire_rate":  1,
			// "body_damage":  10,
			"bullet_speed": 10,
			"reload_speed": 3,
		},
		lastShot: time.Now(),
	}
	log.Printf("Добавлен игрок %d", id)
	return nil
}

func (g *Game) InitObjectSystem(maxObjects int, respawnDelay time.Duration) {
	g.MaxObjects = maxObjects
	g.RespawnDelay = respawnDelay
	g.CheckObjects()
}

func (g *Game) Start() {
	if g.Running {
		return
	}
	g.Running = true

	go func() {
		ticker := time.NewTicker(GameTick)
		defer ticker.Stop()

		for {
			select {
			case <-g.Done:
				return
			case input := <-g.Inputs:
				g.handleInput(input)
			case <-ticker.C:
				g.update()
				g.broadcastState()
				g.updateBullets()
			}
		}
	}()
}

// Stop останавливает игровой цикл
func (g *Game) Stop() {
	if g.Running {
		g.CleanupObjects()
		close(g.Done)
	}
}

func (g *Game) handleInput(input PlayerInput) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	player, ok := g.Players[input.ID]
	if !ok {
		return
	}

	if input.Input.UpgradeStat != "" {
		log.Printf("Стат для апдейта %s", input.Input.UpgradeStat)
		g.handleUpgrade(input.ID, input.Input.UpgradeStat)
	}

	speed := player.Stats["speed"]

	// Предварительные координаты
	newX, newY := player.X, player.Y

	if input.Input.Up {
		newY -= speed
	}
	if input.Input.Down {
		newY += speed
	}
	if input.Input.Left {
		newX -= speed
	}
	if input.Input.Right {
		newX += speed
	}

	// Применяем ограничения
	if newX < MinX {
		newX = MinX
	}
	if newX > MaxX {
		newX = MaxX
	}
	if newY < MinY {
		newY = MinY
	}
	if newY > MaxY {
		newY = MaxY
	}

	// Обновляем координаты
	player.X = newX
	player.Y = newY
	player.Angle = input.Input.Angle

	if input.Input.Shoot {
		g.shootBullet(player)
	}
}

func (g *Game) shootBullet(player *Player) {
	if !player.Alive {
		return
	}

	now := time.Now()
	fireRate := player.Stats["fire_rate"]

	// Безопасный расчет интервала
	if fireRate > 0 {
		minInterval := time.Duration(float64(time.Second) / fireRate)
		if now.Sub(player.lastShot) < minInterval {
			return
		}
	} else {
		return // Если скорострельность нулевая - не стреляем
	}

	bullet := &Bullet{
		ID:        uint(len(g.Bullets) + 1),
		OwnerID:   player.ID,
		X:         player.X,
		Y:         player.Y,
		Angle:     player.Angle,
		Speed:     player.Stats["bullet_speed"],
		Damage:    player.Stats["damage"],
		Active:    true,
		CreatedAt: now,
	}

	player.lastShot = now // Обновляем время последнего выстрела
	g.Bullets = append(g.Bullets, bullet)

	log.Printf("Игрок %d выстрелил. Урон: %.1f, Скорость: %.1f",
		player.ID, bullet.Damage, bullet.Speed)
}

func (g *Game) updateBullets() {
	var activeBullets []*Bullet
	now := time.Now()

	for _, bullet := range g.Bullets {
		if !bullet.Active {
			continue
		}

		// Обновляем позицию пули
		bullet.X += bullet.Speed * math.Cos(bullet.Angle)
		bullet.Y += bullet.Speed * math.Sin(bullet.Angle)

		// Проверяем коллизии
		if !g.checkBulletCollisions(bullet) || !g.checkBulletObjectCollisions(bullet) {
			continue
		}

		// Проверяем границы карты
		if bullet.X < MinX || bullet.X > MaxX || bullet.Y < MinY || bullet.Y > MaxY {
			bullet.Active = false
			continue
		}

		// Проверяем время жизни пули (если нужно)
		if now.Sub(bullet.CreatedAt) > 3*time.Second {
			bullet.Active = false
			continue
		}

		activeBullets = append(activeBullets, bullet)
	}
	g.Bullets = activeBullets
}

func (g *Game) update() {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	for _, player := range g.Players {
		err := player.Conn.WriteJSON(g.serializeState())
		if err != nil {
			log.Printf("ошибка отправки данных игроку %d: %v", player.ID, err)
		}
	}
}

func (g *Game) serializeState() interface{} {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	// Объединяем все игровые сущности в единый ответ
	return map[string]interface{}{
		"players": g.serializePlayers(),
		"bullets": g.serializeBullets(),
		"objects": g.serializeObjects(), // Добавляем игровые объекты
		// "meta": map[string]interface{}{
		// 	"server_time": time.Now().UnixMilli(),
		// 	// "version":     g.Config.Version,
		// },
	}
}

func (g *Game) serializePlayers() map[uint]playerState {
	players := make(map[uint]playerState)
	for id, p := range g.Players {
		players[id] = playerState{
			ID:          p.ID,
			X:           p.X,
			Y:           p.Y,
			Angle:       p.Angle,
			Level:       p.Level,
			Stats:       p.Stats,
			SkillPoints: p.SkillPoints,
		}
	}
	return players
}

func (g *Game) serializeBullets() []bulletState {
	bullets := make([]bulletState, 0, len(g.Bullets))
	for _, b := range g.Bullets {
		bullets = append(bullets, bulletState{
			X:     b.X,
			Y:     b.Y,
			Angle: b.Angle,
		})
	}
	return bullets
}

func (g *Game) serializeObjects() []objectState {
	objects := make([]objectState, 0, len(g.Objects))
	for _, obj := range g.Objects {
		if obj.Active {
			objects = append(objects, objectState{
				ID:     obj.ID,
				X:      obj.X,
				Y:      obj.Y,
				Health: obj.Health,
			})
		}
	}
	return objects
}

// func (g *Game) serializeState() interface{} {

// 	players := make(map[uint]playerState)
// 	for id, p := range g.Players {
// 		players[id] = playerState{
// 			ID:          p.ID,
// 			X:           p.X,
// 			Y:           p.Y,
// 			Angle:       p.Angle,
// 			Level:       p.Level,
// 			Stats:       p.Stats,
// 			SkillPoints: p.SkillPoints,
// 		}
// 	}

// 	bullets := make([]bulletState, 0, len(g.Bullets))
// 	for _, b := range g.Bullets {
// 		bullets = append(bullets, bulletState{
// 			X:     b.X,
// 			Y:     b.Y,
// 			Angle: b.Angle,
// 		})
// 	}

// 	return map[string]interface{}{
// 		"players": players,
// 		"bullets": bullets,
// 	}
// }

func (g *Game) broadcastState() {
	g.Mutex.RLock()
	state := struct {
		Players map[uint]*Player `json:"players"`
		Bullets []*Bullet        `json:"bullets"`
	}{
		Players: g.Players,
		Bullets: g.Bullets,
	}

	var playersToRemove []uint
	players := make([]*Player, 0, len(g.Players))

	// Создаем копию игроков для безопасной итерации
	for _, p := range g.Players {
		players = append(players, p)
	}
	g.Mutex.RUnlock()

	// Проверяем соединения и считаем ошибки
	for _, p := range players {
		if p.Conn == nil {
			playersToRemove = append(playersToRemove, p.ID)
			continue
		}

		// Проверяем состояние соединения перед отправкой
		if err := p.Conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			p.FailedBroadcasts++
		} else {
			if err := p.Conn.WriteJSON(state); err != nil {
				log.Printf("❌ Ошибка отправки состояния игроку %d: %v", p.ID, err)
				p.FailedBroadcasts++
			} else {
				p.FailedBroadcasts = 0 // Сброс при успешной отправке
			}
		}

		if p.FailedBroadcasts >= 3 {
			playersToRemove = append(playersToRemove, p.ID)
		}
	}

	// Безопасное удаление игроков с полной блокировкой
	if len(playersToRemove) > 0 {
		g.Mutex.Lock()
		defer g.Mutex.Unlock()

		for _, id := range playersToRemove {
			if player, exists := g.Players[id]; exists {
				// Дополнительная проверка состояния соединения
				if player.Conn != nil {
					player.Conn.Close()
				}
				delete(g.Players, id)
				log.Printf("⚠️ Игрок %d удален после %d неудачных попыток отправки",
					id, player.FailedBroadcasts)
			}
		}
	}
}

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
