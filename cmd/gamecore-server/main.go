package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"gameCore/internal/bootstrap"
)

func main() {
	// Инициализируем сервисы
	gameInstance, _, router := bootstrap.Init()

	// Обслуживание статических файлов
	// router.Static("/public", "./public")

	// Запускаем игровые процессы
	go gameInstance.Start()

	// Настраиваем и запускаем HTTP сервер
	server := &http.Server{
		Addr:    ":8080",
		Handler: router, // Используем Gin router как основной обработчик
	}

	go func() {
		log.Println("Сервер запущен на http://localhost:8080")
		log.Println("WebSocket endpoint: ws://localhost:8080/ws")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Грейсфул шатдаун
	log.Println("Завершаем работу...")
	// При необходимости добавьте shutdown логику
	// wsServer.Shutdown(context.Background())
	// server.Shutdown(context.Background())
	log.Println("Сервер остановлен.")
}
