package game

import (
	"math"
	"time"
)

type GameObject struct {
	ID         string  // Уникальный идентификатор объекта
	Type       string  // Тип фигуры (круг, квадрат и т. д.)
	Health     int     // Количество HP у объекта
	Experience int     // Сколько опыта дает за уничтожение
	X, Y       float64 // Координаты на карте
}

// Bullet — структура пули с дополнительными полями для времени жизни
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

func (p *Player) DestroyObject(obj *GameObject) {
	if obj.Health <= 0 {
		p.GainXP(obj.Experience)
	}
}

// updateBullets обрабатывает движение и коллизии пуль
func (g *Game) updateBullets() {
	var activeBullets []*Bullet
	now := time.Now()

	for _, bullet := range g.Bullets {
		if !bullet.Active {
			continue
		}

		// Время жизни
		if now.Sub(bullet.CreatedAt) > BulletLifetime {
			bullet.Active = false
			continue
		}

		// Обновляем позицию пули
		bullet.X += bullet.Speed * math.Cos(bullet.Angle)
		bullet.Y += bullet.Speed * math.Sin(bullet.Angle)

		// Проверяем коллизии
		if g.checkBulletCollisions(bullet) {
			activeBullets = append(activeBullets, bullet)
		}
	}
	g.Bullets = activeBullets
}

// shootBullet создает новую пулю
func (g *Game) shootBullet(player *Player) {
	bullet := &Bullet{
		ID:        uint(len(g.Bullets) + 1),
		OwnerID:   player.ID,
		X:         player.X,
		Y:         player.Y,
		Angle:     player.Angle,
		Speed:     player.Stats["bullet_speed"],
		Damage:    player.Stats["damage"],
		Active:    true,
		CreatedAt: time.Now(),
	}
	g.Bullets = append(g.Bullets, bullet)
}
