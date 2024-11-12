package db

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"time"
)

var dbMap = make(map[string]*sqlx.DB)

func DB() *sqlx.DB {
	return dbMap["analyst"]
}
func InitializeDB(config DatabaseConfig) {
	if _, exists := dbMap[config.Name]; exists == true {
		log.Printf("database '%s' is already initialized \n", config.Name)
		return
	}

	dbMap[config.Name] = connectDB(config, 10)
	log.Println("Database connection established")
}

func connectDB(config DatabaseConfig, max int) *sqlx.DB {
	if max >= 0 {
		log.Println("Connect Database", config.Name)

		ssl := "sslmode=require"
		if config.DisableSSL {
			ssl = "sslmode=disable"
		}

		configStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s %s",
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Name,
			ssl)

		db, err := sqlx.Connect("postgres", configStr)

		if err != nil {
			log.Printf("Error connecting the database %s , error: %s\n", config.Name, err)
		} else {
			db.SetMaxOpenConns(50)
			db.SetMaxIdleConns(30)
			db.SetConnMaxIdleTime(5 * time.Minute)
			return db
		}

	} else {
		log.Println("Giving up, can't connect to the database", config.Name)
		os.Exit(0)
	}
	time.Sleep(2 * time.Second)
	log.Printf("try %d to connect to %s\n", max-1, config.Name)
	return connectDB(config, max-1)
}
