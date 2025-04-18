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

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization token required",
			})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

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
	if strings.HasSuffix(c.Request.URL.Path, "/ws") {
		if token := c.Query("token"); token != "" {
			return token
		}
	}

	bearerToken := c.GetHeader("Authorization")
	if len(bearerToken) > 7 && strings.EqualFold(bearerToken[:7], "BEARER ") {
		return bearerToken[7:]
	}

	return ""
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
