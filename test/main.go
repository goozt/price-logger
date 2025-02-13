package main

import (
	"dilogger/internal/db"
	"dilogger/internal/utils"
	"fmt"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	conn := db.ConnectDB(utils.GetEnv("DB_TABLE", "dilogger"))
	ShowData(conn)
}

// The function `ShowData` displays the items in a database connection and the total count of items if there are any.
func ShowData(conn db.Connection) {
	if item_count := len(conn.GetList()); item_count > 0 {
		for _, v := range conn.GetList() {
			fmt.Println(v)
		}
		fmt.Printf("Found %d items\n", item_count)
	}
}
