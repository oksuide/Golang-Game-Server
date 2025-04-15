package game

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"
)

var objectIDCounter uint32

func NewObjectID() uint {
	return uint(atomic.AddUint32(&objectIDCounter, 1))
}

type Object struct {
	ID           uint
	X, Y         float64
	Type         string
	Health       int
	XP           int
	Active       bool
	respawnTimer *time.Timer
}

func (g *Game) AddObject() {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	if len(g.Objects) >= g.MaxObjects {
		return
	}

	object := &Object{
		ID:     NewObjectID(),
		X:      float64(rand.Intn(MaxX-MinX) + MinX),
		Y:      float64(rand.Intn(MaxY-MinY) + MinY),
		Health: 100,
		XP:     50,
		Active: true,
	}

	g.Objects = append(g.Objects, object)
	log.Printf("Создан объект %d (%.1f, %.1f)", object.ID, object.X, object.Y)
}

func (o *Object) Destroy(g *Game, attackerID uint) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()

	o.Active = false
	if attacker, exists := g.Players[attackerID]; exists {
		attacker.GainXP(o.XP)
	}

	o.respawnTimer = time.AfterFunc(g.RespawnDelay, func() {
		g.Mutex.Lock()
		o.Respawn(g)
		g.Mutex.Unlock()
	})
}

func (o *Object) Respawn(g *Game) {
	o.X = float64(rand.Intn(MaxX))
	o.Y = float64(rand.Intn(MaxY))
	o.Active = true
	log.Printf("Объект %d восстановлен", o.ID)
}

func (g *Game) CheckObjects() {
	activeCount := 0
	for _, obj := range g.Objects {
		if obj.Active {
			activeCount++
		}
	}

	if activeCount < g.MaxObjects {
		g.AddObject()
	}
}

func (g *Game) checkBulletObjectCollisions(bullet *Bullet) bool {
	g.Mutex.RLock()
	defer g.Mutex.RUnlock()

	if !bullet.Active {
		return false
	}

	for _, obj := range g.Objects {
		if !obj.Active {
			continue
		}

		dx := bullet.X - obj.X
		dy := bullet.Y - obj.Y
		if dx*dx+dy*dy < (ObjectRadius+BulletRadius)*(ObjectRadius+BulletRadius) {
			obj.ObjectTakeDamage(int(bullet.Damage), g, bullet.OwnerID)
			bullet.Active = false
			return false
		}
	}
	return true
}

func (o *Object) ObjectTakeDamage(damage int, game *Game, attackerID uint) {
	o.Health -= damage
	log.Printf("Игрок %d получил %d урона от %d. Осталось здоровья: %d",
		o.ID, damage, attackerID, o.Health)

	if o.Health <= 0 {
		o.Destroy(game, attackerID)
	}
}

func (g *Game) CleanupObjects() {
	for _, obj := range g.Objects {
		if obj.respawnTimer != nil {
			obj.respawnTimer.Stop()
		}
	}
	g.Objects = nil
}
