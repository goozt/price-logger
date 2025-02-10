package main

import (
	"context"
	"dilogger/internal/db"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

//go:embed static
var tplFolder embed.FS

var conn = db.ConnectDB("dilogger")

// Get unique product names
func getUniqueProducts() []string {
	unique := make(map[string]bool)
	var productList []string
	for _, p := range conn.GetList() {
		if !unique[p.Name] {
			unique[p.Name] = true
			productList = append(productList, p.Name)
		}
	}
	return productList
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "static/index.html")
	if err != nil {
		log.Fatal(err)
	}
	println("Embedded the index.html file")
	tmpl.Execute(w, nil)
}

func resetHandler(w http.ResponseWriter, r *http.Request) {
	conn.ResetDB()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{ Reset bool }{true})
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getUniqueProducts())
}

func pricesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	var priceData []db.Product
	for _, p := range conn.GetList() {
		if p.Name == name {
			priceData = append(priceData, p)
		}
	}
	json.NewEncoder(w).Encode(priceData)
}

func handleShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
}
