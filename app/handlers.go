package main

import (
	"dilogger/internal/parser"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
)

//go:embed static
var tplFolder embed.FS

// The indexHandler function parses and executes an embedded index.html template for a web server response.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "static/index.html")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, nil)
}

// The indexHandler function parses and executes an embedded index.html template for a web server response.
func jsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "static/main.js")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/javascript")
	tmpl.Execute(w, nil)
}

// The indexHandler function parses and executes an embedded index.html template for a web server response.
func cssHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "static/main.css")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/css")
	tmpl.Execute(w, nil)
}

// The `newDataHandler` function saves data to a database and responds with a JSON object indicating new data.
func newDataHandler(w http.ResponseWriter, r *http.Request) {
	parser.SaveToDB(conn, wishlist_urls)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{ NewData bool }{true})
}

// The `resetHandler` function resets the database and returns a JSON response indicating the reset operation.
func resetHandler(w http.ResponseWriter, r *http.Request) {
	conn.ResetDB()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{ Reset bool }{true})
}

// The productsHandler function sets the content type to JSON and encodes unique products to be sent in the response.
func productsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conn.GetUniqueNameList())
}

// The `pricesHandler` function retrieves and returns price data for a specific product name in JSON format.
func pricesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name")
	json.NewEncoder(w).Encode(conn.GetListByName(name))
}
