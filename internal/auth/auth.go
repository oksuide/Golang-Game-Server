package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"gameCore/internal/config"
	"gameCore/internal/repository"
	"gameCore/internal/storage"
	"gameCore/pkg/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func RegisterUser(ctx context.Context, repo repository.UserRepository, username, password, email string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	emailExists, err := repo.EmailExists(ctx, email)
	if err != nil {
		return fmt.Errorf("check email exists: %w", err)
	}
	if emailExists {
		return storage.ErrEmailAlreadyExists
	}

	exists, err := repo.UserExists(ctx, username)
	if err != nil {
		return fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return storage.ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	if err := repo.CreateUser(ctx, &user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func LoginUser(ctx context.Context, repo repository.UserRepository, cfg *config.Config, username, password string) (string, error) {
	user, err := repo.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", storage.ErrInvalidCredentials
	}

	claims := &Claims{
		UserID:   user.ID, // добавляем ID
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWT.Expiration)),
		},
	}

	log.Printf("Token claims: %+v", claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.SecretKey))
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return tokenString, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return storage.ErrInvalidPassword
	}
	return nil
}

type AuthHandler struct {
	repo repository.UserRepository
	cfg  *config.Config
}

func NewAuthHandler(repo repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		repo: repo,
		cfg:  cfg,
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := RegisterUser(c.Request.Context(), h.repo, req.Username, req.Password, req.Email)
	switch {
	case errors.Is(err, storage.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
	case errors.Is(err, storage.ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password requirements: minimum 8 characters",
		})
	case errors.Is(err, storage.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})

	case err != nil:
		log.Printf("Registration error: %v", err) // Логируем детали ошибки
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "registration failed",
			"details": "check server logs", // Для разработки
		})
	default:
		// После успешной регистрации сразу логиним
		token, err := LoginUser(c.Request.Context(), h.repo, h.cfg, req.Username, req.Password)
		if err != nil {
			log.Printf("Auto-login after registration failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auto-login failed"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"token":      token,
			"expires_in": h.cfg.JWT.Expiration / time.Second,
		})
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := LoginUser(c.Request.Context(), h.repo, h.cfg, req.Username, req.Password)
	switch {
	case errors.Is(err, storage.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
	default:
		c.JSON(http.StatusOK, gin.H{
			"token":      token,
			"expires_in": h.cfg.JWT.Expiration / time.Second,
		})
	}
}
