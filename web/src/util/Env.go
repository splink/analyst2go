package util

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

func LoadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
}

func Env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}
