package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	// Роут для WebSocket
	// router.GET("/ws", gin.WrapH(http.HandlerFunc(websocket.HandleWebSocket)))

	// Роуты API
	// api := router.Group("/api")

	// Роуты без авторизации
	// api.GET("/rooms/events", handlers.GetRoomEvents)
	// api.POST("/register", handlers.CreateUser)
	// api.POST("/login", handlers.Login)

	// Роуты с авторизацией
	// authorized := api.Group("")
	// 	authorized.Use(middleware.AuthMiddleware())
	// 	{
	// 		setupUserRoutes(authorized)
	// 	}

	return router
}
