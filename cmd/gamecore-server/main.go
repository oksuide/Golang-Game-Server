package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gameCore/internal/bootstrap"
)

func main() {
	// Инициализируем сервисы
	gameInstance, wsServer := bootstrap.Init()

	// Запускаем игровые процессы
	go gameInstance.Start()
	go wsServer.Start(":8080")

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Грейсфул шатдаун
	log.Println("Завершаем работу...")
	ctx := context.Background()
	wsServer.Shutdown(ctx)
	log.Println("Сервер остановлен.")
}
