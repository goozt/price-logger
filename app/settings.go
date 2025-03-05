package app

import (
	"dilogger/internal/db"
	"dilogger/internal/utils"
	"io/fs"
	"slices"

	"github.com/pocketbase/pocketbase/core"
)

// Customise settings
func InitSettings(app core.App) {
	settings := app.Settings()
	settings.Meta.AppName = utils.GetEnv("PB_APP_NAME", "Price Logger")
	settings.Meta.AppURL = utils.GetEnv("PB_APP_URL", "http://localhost:8090")
	settings.Logs.MaxDays = 7
	settings.Logs.LogAuthId = false
	settings.Logs.LogIP = false
	settings.Logs.MinLevel = -4
	err := app.Save(settings)
	if err != nil {
		app.Logger().Error(err.Error())
	}
}

// Add all routes onServe trigger
func AddRoutes(s *db.Server, htmlFS fs.FS, staticFS fs.FS) {
	s.OnServe().BindFunc(func(se *core.ServeEvent) error {
		AddStopRoute(se)
		AddUIRoute(se, htmlFS)
		AddStaticRoute(se, staticFS)
		AddReloadRoute(se, s)
		return se.Next()
	})
}

// Add cron jobs
func AddHourlyJob(s *db.Server, id string, Job func()) {
	cron := s.Cron()
	cron.MustAdd(id, "0 * * * *", func() { Job() })
	cron.Start()
}

// Add monitoring functions
func AddMonitor(s *db.Server) {
	s.PriceUpdateHook(func(e *core.RecordEvent) error {
		if s.CountPriceRecords(e.Record.GetString("product")) > 1 {
			product := s.GetProduct(e.Record)
			if product.Id != "" {
				s.Notification.Send(product)
			}
		}
		return e.Next()
	})
}

// Add intial set of urls to database
func AddURL(app core.App, url string, isProductUrl ...bool) {
	collection, err := app.FindCollectionByNameOrId("urls")
	if err != nil {
		app.Logger().Error(err.Error())
		return
	}
	productUrlType := "wishlist"
	if len(isProductUrl) > 0 {
		if isProductUrl[0] {
			productUrlType = "product"
		}
	}
	record := core.NewRecord(collection)
	record.Set("url", url)
	record.Set("type", productUrlType)
	err = app.Save(record)
	if err != nil {
		app.Logger().Error(err.Error())
	}
}

// Add user to login
func AddUser(app core.App, userdb string) {
	collection, err := app.FindCollectionByNameOrId(userdb)
	if err != nil {
		app.Logger().Error(err.Error())
		return
	}
	record := core.NewRecord(collection)
	fields := record.Collection().Fields.FieldNames()
	if slices.Contains(fields, "name") {
		record.Set("name", utils.GetEnv("PB_ADMIN_NAME", "Super Admin"))
		record.SetVerified(true)
	}
	record.SetEmail(utils.GetEnv("PB_ADMIN_EMAIL"))
	record.SetPassword(utils.GetEnv("PB_ADMIN_PASSWORD"))
	err = app.Save(record)
	if err != nil {
		app.Logger().Error(err.Error())
	}
}
