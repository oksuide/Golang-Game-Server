package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gameCore/config"
	"gameCore/pkg/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type UserRepository interface {
	UserExists(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, username string) (*models.User, error)
}

func RegisterUser(ctx context.Context, repo UserRepository, username, password string) error {
	if err := validatePassword(password); err != nil {
		return err
	}

	exists, err := repo.UserExists(ctx, username)
	if err != nil {
		return fmt.Errorf("check user exists: %w", err)
	}
	if exists {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := models.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := repo.CreateUser(ctx, &user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func LoginUser(ctx context.Context, repo UserRepository, cfg *config.Config, username, password string) (string, error) {
	user, err := repo.GetUser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	claims := &Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.JWT.Expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.JWT.SecretKey))
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return tokenString, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}
