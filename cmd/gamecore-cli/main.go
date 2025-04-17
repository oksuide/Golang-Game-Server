package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

const (
	serverURL  = "ws://localhost:8080/ws"
	maxRetries = 3
	retryDelay = 2 * time.Second
)

func main() {
	// 1. Добавляем получение JWT токена из переменных окружения
	token := os.Getenv("GAMECORE_TOKEN")
	if token == "" {
		log.Fatal("GAMECORE_TOKEN не установлен. Пример:\nexport GAMECORE_TOKEN='ваш_jwt_токен'")
	}

	// 2. Настраиваем заголовки для аутентификации
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)

	var conn *websocket.Conn
	var err error

	// 3. Добавляем повторные попытки подключения
	for attempt := 1; attempt <= maxRetries; attempt++ {
		dialer := &websocket.Dialer{
			HandshakeTimeout: 5 * time.Second,
			Proxy:            http.ProxyFromEnvironment,
			TLSClientConfig:  websocket.DefaultDialer.TLSClientConfig,
		}

		conn, _, err = dialer.Dial(serverURL, headers)
		if err == nil {
			break
		}

		log.Printf("Попытка %d/%d: %v", attempt, maxRetries, err)
		time.Sleep(retryDelay)
	}

	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()
	log.Println("✅ Успешное подключение к серверу!")

	// 4. Улучшаем обработку прерываний
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// 5. Добавляем буферизированный канал для ввода
	inputChan := make(chan string, 10)

	// Горутина для чтения ввода
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
		close(inputChan)
	}()

	// 6. Улучшенная обработка сообщений от сервера
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("❌ Ошибка чтения: %v", err)
				}
				return
			}
			log.Printf("📩 [Сервер]: %s\n", string(message))
		}
	}()

	log.Println("Введите команду (help - список команд):")

loop:
	for {
		select {
		case <-done:
			break loop
		case <-interrupt:
			log.Println("Получен сигнал прерывания...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Ошибка при закрытии соединения:", err)
			}
			break loop
		case cmd, ok := <-inputChan:
			if !ok {
				break loop
			}
			handleCommand(cmd, conn, interrupt)
		}
	}

	log.Println("CLI завершил работу")
}

// 7. Выносим обработку команд в отдельную функцию
func handleCommand(cmd string, conn *websocket.Conn, interrupt chan<- os.Signal) {
	switch cmd {
	case "help":
		log.Println("Доступные команды:")
		log.Println("help - показать команды")
		log.Println("exit - выход")
		log.Println("send <message> - отправить сообщение")
	case "exit":
		interrupt <- os.Interrupt
	default:
		if err := conn.WriteMessage(websocket.TextMessage, []byte(cmd)); err != nil {
			log.Println("❌ Ошибка отправки:", err)
		} else {
			log.Println("📨 Сообщение отправлено")
		}
	}
}
