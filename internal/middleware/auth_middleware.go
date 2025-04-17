package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig interface {
	GetJWTSecret() string
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение конфигурации
		cfg, ok := c.MustGet("config").(JWTConfig)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Configuration unavailable",
			})
			return
		}

		// Извлечение токена
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization required",
			})
			return
		}

		// Парсинг токена
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.GetJWTSecret()), nil
		})

		// Обработка ошибок парсинга
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Проверка валидности токена
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		// Извлечение claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		// Извлечение userID
		userID, err := extractUserID(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.EqualFold(bearerToken[:7], "BEARER ") {
		return bearerToken[7:]
	}
	return c.Query("token")
}

func extractUserID(claims jwt.MapClaims) (uint, error) {
	claimValue, ok := claims["user_id"]
	if !ok {
		return 0, fmt.Errorf("токен не содержит user_id")
	}

	switch v := claimValue.(type) {
	case float64:
		return uint(v), nil
	case string:
		// Обработка строкового представления ID
		var id uint
		_, err := fmt.Sscanf(v, "%d", &id)
		if err != nil {
			return 0, fmt.Errorf("некорректный формат user_id")
		}
		return id, nil
	default:
		return 0, fmt.Errorf("неподдерживаемый тип user_id")
	}
}
