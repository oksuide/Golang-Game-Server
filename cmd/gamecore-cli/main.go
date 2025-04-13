package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

const serverURL = "ws://localhost:8080/ws"

func main() {
	// Подключаемся к серверу
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		log.Fatalf("Ошибка подключения к серверу: %v", err)
	}
	defer conn.Close()
	log.Println("✅ Подключен к серверу!")

	// Канал для обработки прерываний (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Запускаем горутину для чтения сообщений от сервера
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("❌ Ошибка чтения:", err)
				return
			}
			log.Println("📩 Получено сообщение:", string(message))
			// В этой части можно обрабатывать данные и передавать их на фронт для визуализации
		}
	}()

	// Цикл отправки сообщений
	for {
		select {
		case <-interrupt:
			log.Println("Отключение от сервера...")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}
