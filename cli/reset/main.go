package main

import "dilogger/internal/db"

// The main function connects to a database named "dilogger" and resets it.
func main() {
	conn := db.ConnectDB("dilogger")
	conn.ResetDB()
}
