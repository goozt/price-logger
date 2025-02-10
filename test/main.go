package main

import (
	"dilogger/internal/db"
	"fmt"
)

func main() {
	conn := db.ConnectDB("dilogger")
	ShowData(conn)
}

func ShowData(conn db.Connection) {
	if item_count := len(conn.GetList()); item_count > 0 {
		for _, v := range conn.GetList() {
			fmt.Println(v)
		}
		fmt.Printf("Found %d items\n", item_count)
	}
}
