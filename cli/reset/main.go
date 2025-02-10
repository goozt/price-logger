package main

import "dilogger/internal/db"

func main() {
	conn := db.ConnectDB("dilogger")
	conn.ResetDB()
}
