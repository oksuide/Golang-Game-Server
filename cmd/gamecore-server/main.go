package main

import (
	"log"
	"net/http"

	"gameCore/internal/bootstrap"
)

func main() {
	// Инициализируем сервисы (БД, Redis, движок, WebSocket)
	gameInstance, wsServer := bootstrap.Init()

	// Запускаем игровой цикл
	go gameInstance.GameLoop()
	go wsServer.BroadcastGameState()

	// Настроим маршруты
	http.HandleFunc("/ws", wsServer.HandleConnections)

	// Запускаем сервер
	port := 8080
	log.Printf("Сервер запущен на порту :%d", port)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Ошибка сервера:", err)
	}
}
