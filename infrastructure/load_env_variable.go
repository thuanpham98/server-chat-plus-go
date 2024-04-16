package infrastructure

import (
	"log"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnvVariable(){
	err:= godotenv.Load(filepath.Join("", ".env"))
	if(err!=nil){
		log.Fatalf("Error when loading .env : %v",err)
	}
}