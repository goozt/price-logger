package main

import (
	"dilogger/internal/db"
	"dilogger/internal/parser"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"text/template"
)

//go:embed static
var tplFolder embed.FS

var conn = db.ConnectDB("dilogger")

// The function `getUniqueProducts` returns a list of unique product names by filtering out duplicates.
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

// The indexHandler function parses and executes an embedded index.html template for a web server response.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(tplFolder, "static/index.html")
	if err != nil {
		log.Fatal(err)
	}
	println("Embedded the index.html file")
	tmpl.Execute(w, nil)
}

// The `newDataHandler` function saves data to a database and responds with a JSON object indicating new data.
func newDataHandler(w http.ResponseWriter, r *http.Request) {
	parser.SaveToDB(wishlist_urls)
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
	json.NewEncoder(w).Encode(getUniqueProducts())
}

// The `pricesHandler` function retrieves and returns price data for a specific product name in JSON format.
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
