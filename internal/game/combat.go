package game

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// TakeDamage обрабатывает получение урона игроком
func (p *Player) TakeDamage(damage float64, game *Game, attackerID uint) {
	p.Stats["health"] -= damage
	log.Printf("Игрок %d получил %.1f урона от %d. Осталось здоровья: %.1f",
		p.ID, damage, attackerID, p.Stats["health"])

	if p.Stats["health"] <= 0 {
		p.Die(game, attackerID)
	}
}

// checkBulletCollisions проверяет коллизии пули с игроками
func (g *Game) checkBulletCollisions(bullet *Bullet) bool {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	for _, player := range g.Players {
		// Пуля не может попасть в своего владельца или мертвого игрока
		if player.ID == bullet.OwnerID || !player.Alive {
			continue
		}

		if g.isColliding(bullet, player) {
			player.TakeDamage(bullet.Damage, g, bullet.OwnerID)
			bullet.Active = false
			return false
		}
	}
	return true
}

func (g *Game) isColliding(bullet *Bullet, player *Player) bool {
	dx := bullet.X - player.X
	dy := bullet.Y - player.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < (PlayerRadius + BulletRadius)
}

func (p *Player) Die(game *Game, killerID uint) {
	p.Alive = false
	log.Printf("Игрок %d убит игроком %d", p.ID, killerID)

	// Запускаем таймер респавна
	p.RespawnTimer = time.AfterFunc(RespawnTime, func() {
		game.RespawnPlayer(p.ID)
	})

	// Обрабатываем убийство
	game.HandleKill(killerID, p.ID)
}

func (g *Game) RespawnPlayer(playerID uint) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if player, exists := g.Players[playerID]; exists {
		player.Alive = true
		player.Stats["health"] = BasePlayerHealth
		player.X = float64(rand.Intn(MaxX))
		player.Y = float64(rand.Intn(MaxY))
		log.Printf("Игрок %d возродился", playerID)
	}
}

func (g *Game) HandleKill(killerID, victimID uint) {
	killer, kExists := g.Players[killerID]
	victim, vExists := g.Players[victimID]

	if !kExists || !vExists {
		return
	}

	lostXP := 0
	gainedXP := 300
	if lostXP > 0 {
		victim.XP -= lostXP
	}
	if gainedXP > 0 {
		killer.GainXP(gainedXP)
	}
	log.Printf("Игрок %d убил игрока %d", killerID, victimID)
}

func (g *Player) GainXP(gainedXP int) {
	if gainedXP > 0 {
		g.XP += gainedXP
		log.Printf("Игрок %d получил %d опыта, текущий опыт %d", g.ID, gainedXP, g.XP)
		g.CheckLvlUp()
	}
}

func (g *Player) CheckLvlUp() {
	leveledUp := false

	for g.XP >= g.NewLvlExp {
		g.Level += 1
		g.SkillPoints += 1
		leveledUp = true
		log.Printf("Игрок %d повысил уровень! Текущий уровень %d. Доступные очки прокачки %d",
			g.ID, g.Level, g.SkillPoints)
		g.NewLvlExp = g.Level * g.Level * 100
	}
	if leveledUp {
		// Отправляем обновлённые данные игроку
		g.Conn.WriteJSON(map[string]interface{}{
			"skill_points": g.SkillPoints,
			"level":        g.Level,
			"new_lvl_exp":  g.NewLvlExp,
		})
	}

	log.Printf("Кап до нового уровня: %d", (g.NewLvlExp))
}

func (g *Game) handleUpgrade(playerID uint, stat string) {
	player, exists := g.Players[playerID]
	if !exists || player.SkillPoints <= 0 {
		return
	}

	// Применяем улучшение
	switch stat {
	case "damage":
		player.Stats["damage"] += 20
		log.Printf("Урон увеличен до: %.1f", player.Stats["damage"])
	case "health":
		player.Stats["health"] += 15
		player.Stats["max_health"] += 15
		log.Printf("Здоровье увеличено до: %.1f", player.Stats["health"])
	case "speed":
		if player.Stats["speed"] < 15 {
			player.Stats["speed"] += 0.5
			log.Printf("Скорость увеличена до: %.1f", player.Stats["speed"])
		}
	default:
		log.Printf("Неизвестный стат для улучшения: %s", stat)
		return
	}

	player.SkillPoints--

	// Асинхронная отправка обновлений
	go g.sendPlayerUpdate(playerID)

	log.Printf("Игрок %d улучшил %s. Осталось очков: %d",
		playerID, stat, player.SkillPoints)
}

func (g *Game) sendPlayerUpdate(playerID uint) {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	player, exists := g.Players[playerID]
	if !exists || player.Conn == nil {
		return
	}

	err := player.Conn.WriteJSON(map[string]interface{}{
		"type":         "upgrade",
		"skill_points": player.SkillPoints,
		"stats":        player.Stats,
	})

	if err != nil {
		log.Printf("Ошибка отправки обновления игроку %d: %v", playerID, err)
	}
}
