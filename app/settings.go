package app

import (
	"dilogger/internal/db"
	"dilogger/internal/utils"
	"errors"
	"io/fs"
	"net/http"
	"slices"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

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

func AddPaths(s *db.Server, htmlFS fs.FS, staticFS fs.FS, reload func(server *db.Server)) {
	s.OnServe().BindFunc(func(se *core.ServeEvent) error {
		registry := template.NewRegistry()
		se.Router.GET("/", func(e *core.RequestEvent) error {
			html, err := registry.LoadFS(
				htmlFS,
				"pb_public/html/layout.html",
				"pb_public/html/settings.html",
				"pb_public/html/login.html",
				"pb_public/html/index.html",
			).Render(map[string]string{
				"title":  utils.GetEnv("PB_APP_NAME", "Price Logger"),
				"apiUrl": utils.GetEnv("PB_APP_URL", "http://localhost:8090"),
			})
			if err != nil {
				return e.NotFoundError("", err)
			}
			return e.HTML(http.StatusOK, html)
		})
		staticFS, err := fs.Sub(staticFS, "pb_public/static")
		if err != nil {
			return errors.New("error: static directory missing")
		}
		se.Router.GET("/static/{path...}", apis.Static(staticFS, false))
		se.Router.GET("/api/reload-data", func(e *core.RequestEvent) error {
			reload(s)
			return e.JSON(http.StatusOK, map[string]bool{
				"reloaded": true,
			})
		})
		return se.Next()
	})
}

func AddHourlyJob(s *db.Server, id string, Job func()) {
	cron := s.Cron()
	cron.MustAdd(id, "0 * * * *", func() { Job() })
	cron.Start()
}

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
