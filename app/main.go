package main

import (
	"dilogger/internal/db"
	"dilogger/internal/parser"
	"dilogger/internal/utils"
	"errors"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
)

var (
	port *string
	conn db.Connection
)

var wishlist_urls = []string{
	"https://www.designinfo.in/wishlist/view/f6a054/",
	"https://www.designinfo.in/wishlist/view/da0c1e/",
}

// The main function sets up a web server with various handlers and options, including a background process for saving data to a database periodically.
func main() {
	server := InitServer(wishlist_urls)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/main.js", jsHandler)
	http.HandleFunc("/main.css", cssHandler)
	http.HandleFunc("/api/products", productsHandler)
	http.HandleFunc("/api/prices", pricesHandler)
	http.HandleFunc("/api/reset", resetHandler)
	http.HandleFunc("/api/new", newDataHandler)

	shutdownChan := make(chan bool, 1)
	go func() {
		log.Println("Server running on http://localhost:" + *port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		shutdownChan <- true
	}()

	utils.HandleShutdown(server)
	<-shutdownChan
}

// The function `ProductLogging` runs a goroutine that periodically saves data from URLs to a database using a parser and connection, with a time interval of one hour.
func ProductLogging(urls []string) {
	go func() {
		for {
			parser.SaveToDB(conn, urls)
			time.Sleep(time.Hour)
		}
	}()
}

// The InitServer function initializes a server with optional web server functionality and product logging.
func InitServer(urls []string) *http.Server {
	godotenv.Load()
	conn = db.ConnectDB(utils.GetEnv("DB_TABLE", "dilogger"))

	port = flag.String("port", "8888", "Port number")
	web_only := flag.Bool("web", false, "Run only web server")
	flag.Parse()

	if !*web_only {
		ProductLogging(urls)
	}

	return &http.Server{Addr: ":" + *port}
}
