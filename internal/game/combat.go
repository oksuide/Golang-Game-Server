package game

import (
	"log"
	"math"
	"time"
)

// handleShooting обрабатывает выстрелы игрока
func (g *Game) handleShooting(player *Player, input struct {
	Up, Down, Left, Right bool
	Angle                 float64
	Shoot                 bool
}) {
	if !input.Shoot {
		return
	}

	// Проверяем скорострельность
	fireRate := player.Stats["fire_rate"]
	if time.Since(player.lastShot) < time.Duration(1.0/fireRate)*time.Second {
		return
	}

	g.shootBullet(player)
	player.lastShot = time.Now()
}

// HandleKill обрабатывает убийство игрока
func (g *Game) HandleKill(killerID, victimID uint) {
	killer, kExists := g.Players[killerID]
	victim, vExists := g.Players[victimID]
	if !kExists || !vExists {
		return
	}

	lostXP := int(float64(victim.XP) * 0.4)
	gainedXP := int(float64(victim.XP) * 0.2)

	if lostXP > 0 {
		victim.XP -= lostXP
	}
	if gainedXP > 0 {
		killer.GainXP(gainedXP)
	}

	log.Printf("Игрок %d убил %d: получил %d XP, жертва потеряла %d XP", killerID, victimID, gainedXP, lostXP)
	g.RemovePlayer(victimID)
}

// TakeDamage обрабатывает получение урона игроком
func (p *Player) TakeDamage(damage float64, game *Game, attackerID uint) {
	p.Stats["health"] -= damage
	if p.Stats["health"] <= 0 {
		game.HandleKill(attackerID, p.ID)
	}
}

// checkBulletCollisions проверяет коллизии пули с игроками
func (g *Game) checkBulletCollisions(bullet *Bullet) bool {
	for _, player := range g.Players {
		if player.ID == bullet.OwnerID || !bullet.Active {
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

func (p *Player) CollideWith(other *Player, g *Game) {
	// Оба игрока получают урон от столкновения
	p.Stats["health"] -= other.Stats["body_damage"]
	other.Stats["health"] -= p.Stats["body_damage"]

	// Проверка на смерть
	if p.Stats["health"] <= 0 {
		g.HandleKill(other.ID, p.ID)
	}
	if other.Stats["health"] <= 0 {
		g.HandleKill(p.ID, other.ID)
	}
}

// isColliding проверяет коллизию между пулей и игроком
func (g *Game) isColliding(bullet *Bullet, player *Player) bool {
	dx := bullet.X - player.X
	dy := bullet.Y - player.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < CollisionDistance
}
