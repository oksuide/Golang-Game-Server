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
	gameInstance, wsServer := bootstrap.Init()

	// Статические файлы (например, index.html)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web"))))

	// Запускаем игровые процессы
	go gameInstance.Start()
	go wsServer.StartServer()

	// Запускаем HTTP сервер для обслуживания статических файлов
	go func() {
		log.Println("HTTP сервер работает на порту 8081 для статики")
		err := http.ListenAndServe(":8081", nil) // Это другой порт для статики
		if err != nil {
			log.Fatalf("Ошибка HTTP сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Грейсфул шатдаун
	log.Println("Завершаем работу...")
	// wsServer.Shutdown(nil)
	log.Println("Сервер остановлен.")
}
