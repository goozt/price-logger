package app

import (
	"dilogger/internal/db"
	"dilogger/internal/product"
	"dilogger/internal/utils"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

// Add route to stop the process
func AddStopRoute(se *core.ServeEvent) {
	err := os.WriteFile(filepath.Join(se.App.DataDir(), ".pid"), fmt.Append(nil, syscall.Getpid()), 0644)
	if err != nil {
		se.App.Logger().Error(err.Error())
		return
	}
	exit := make(chan bool, 1)
	se.Router.POST("/stop", func(e *core.RequestEvent) error {
		event := new(core.TerminateEvent)
		event.App = se.App
		data := struct {
			Id string `json:"id" form:"id"`
		}{}
		if err := e.BindBody(&data); err != nil {
			return e.BadRequestError("failed to read request data", err)
		}
		pid := fmt.Sprint(syscall.Getpid())
		if data.Id != pid {
			return fmt.Errorf("stop api call failed: invalid app id")
		}
		se.App.OnTerminate().Trigger(event, func(e *core.TerminateEvent) error {
			err := e.App.ResetBootstrapState()
			if err == nil {
				go func() {
					<-exit
					syscall.Exit(0)
				}()
			}
			return nil
		})
		return e.JSON(http.StatusOK, map[string]bool{"success": true})
	})
	exit <- true
}

// Add routes to HTML UI
func AddUIRoute(se *core.ServeEvent, hfs fs.FS) {
	registry := template.NewRegistry()
	se.Router.GET("/", func(e *core.RequestEvent) error {
		html, err := registry.LoadFS(
			hfs,
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
}

// Add static routes
func AddStaticRoute(se *core.ServeEvent, sfs fs.FS) {
	staticFS, err := fs.Sub(sfs, "pb_public/static")
	if err != nil {
		se.App.Logger().Error(err.Error())
		return
	}
	se.Router.GET("/static/{path...}", apis.Static(staticFS, false))
}

// Add route to reload product data
func AddReloadRoute(se *core.ServeEvent, server *db.Server) {
	se.Router.GET("/api/reload-data", func(e *core.RequestEvent) error {
		product.ReloadData(server)
		return e.JSON(http.StatusOK, map[string]bool{
			"reloaded": true,
		})
	})
}
