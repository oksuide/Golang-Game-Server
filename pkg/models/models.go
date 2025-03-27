package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);not null;unique"`
	Email    string `gorm:"type:text;not null;unique"`
	Password string `gorm:"type:text;not null"`
}
