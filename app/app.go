package app

import (
	"dilogger/internal/db"
	"dilogger/internal/product"
	"embed"

	"github.com/pocketbase/pocketbase/core"
)

//go:embed all:pb_public/html
var html embed.FS

//go:embed all:pb_public/static
var static embed.FS

func Run() {
	server := db.NewServer()
	AddMonitor(server)
	AddPaths(server, html, static, product.ReloadData)
	AddHourlyJob(server, "pricelogger", func() {
		product.ReloadData(server)
	})
	server.AddCommand("init", func(app core.App) {
		server.NewUrlCollection()
		server.NewProductCollection()
		server.NewPriceCollection()
		InitSettings(app)
		AddUser(app, "_superusers")
		AddUser(app, "users")
		AddURL(app, "https://www.designinfo.in/wishlist/view/f6a054/")
		AddURL(app, "https://www.designinfo.in/wishlist/view/da0c1e/")
		product.ReloadData(server)
	})
	// AddHourlyJob(server, "pricemonitor", func(s *db.Server) {
	// 	monitor.MonitorPriceChanges(s)
	// })
	server.Start()
}
