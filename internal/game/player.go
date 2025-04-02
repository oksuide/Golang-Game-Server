package game

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID    uint
	X, Y  float64
	Angle float64
	Conn  *websocket.Conn `json:"-"` // Исключаем из JSON, так как соединение не сериализуется

	Level       int
	XP          int
	SkillPoints int
	Stats       map[string]float64

	lastShot time.Time // Время последнего выстрела для контроля скорострельности
}

// handleUpgrades обрабатывает улучшения характеристик
func (g *Game) handleUpgrades(player *Player, upgradeStat string) {
	if upgradeStat != "" {
		player.UpgradeStats(upgradeStat)
	}
}

// LevelUp проверяет и выполняет повышение уровня игрока
func (p *Player) LevelUp() {
	requiredXP := p.Level * p.Level * 100
	if p.XP >= requiredXP {
		p.Level++
		p.SkillPoints++
		log.Printf("Игрок %d достиг уровня %d!", p.ID, p.Level)
	}
}

// UpgradeStats улучшает характеристику игрока
func (p *Player) UpgradeStats(stat string) {
	if p.SkillPoints <= 0 {
		return
	}

	upgradeMap := map[string]float64{
		"health":       10,
		"damage":       2,
		"speed":        0.5,
		"fire_rate":    -0.05, // Уменьшаем время между выстрелами
		"body_damage":  3,
		"bullet_speed": 1,
		"reload_speed": -0.25, // Уменьшаем время перезарядки
	}

	if val, exists := upgradeMap[stat]; exists {
		p.Stats[stat] += val
		p.SkillPoints--
		log.Printf("Игрок %d улучшил %s до %.2f", p.ID, stat, p.Stats[stat])
	}
}

// GainXP добавляет опыт игроку
func (p *Player) GainXP(amount int) {
	if amount <= 0 {
		return
	}

	p.XP += amount
	p.LevelUp() // Проверяем, можно ли повысить уровень
}
