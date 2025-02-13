package main

import (
	"dilogger/internal/db"
	"dilogger/internal/utils"
	"log"

	"github.com/joho/godotenv"
)

// The main function connects to a database named "dilogger" and resets it.
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	conn := db.ConnectDB(utils.GetEnv("DB_TABLE", "dilogger"))
	conn.ResetDB()
}
