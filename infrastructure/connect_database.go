package infrastructure

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB * gorm.DB

func ConnectToDatabase(){
	dsn := os.Getenv("DBCONFIG")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if(err!=nil){
		panic("Connect database falled")
	}
	DB = db
}