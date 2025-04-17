package storage

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrHashingPassword    = errors.New("password hashing failed")
	ErrTokenGeneration    = errors.New("token generation failed")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// Роуты с авторизацией
// authorized := api.Group("")
// 	authorized.Use(middleware.AuthMiddleware())
// 	{
// 		setupUserRoutes(authorized)
// 	}
