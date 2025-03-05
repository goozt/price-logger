package app

import (
	"dilogger/internal/db"
	"dilogger/internal/product"
	"embed"
)

//go:embed all:pb_public/html
var html embed.FS

//go:embed all:pb_public/static
var static embed.FS

func Run() {
	server := db.NewServer()
	AddMonitor(server)
	AddRoutes(server, html, static)
	AddHourlyJob(server, "pricelogger", func() {
		product.ReloadData(server)
	})
	Start(server)
}
