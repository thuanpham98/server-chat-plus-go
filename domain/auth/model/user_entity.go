package domain_auth_model

import "gorm.io/gorm"

type UserEntity struct{
	gorm.Model
	Id   string `gorm:"primaryKey"`
	Name string
	Email string `gorm:"unique"`
	Phone string `gorm:"unique"`
	Password string
}