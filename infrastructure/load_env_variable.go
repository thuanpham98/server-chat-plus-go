package infrastructure

import (
	"log"

	"github.com/joho/godotenv"
)

func LoadEnvVariable(){
	err:= godotenv.Load()
	if(err!=nil){
		log.Fatal("Error when loading .env")
	}
}