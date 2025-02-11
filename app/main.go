package main

import (
	"dilogger/internal/parser"
	"errors"
	"flag"
	"log"
	"net/http"
	"time"
)

var (
	host string
	port string
)

var wishlist_urls = []string{
	"https://www.designinfo.in/wishlist/view/f6a054/",
	"https://www.designinfo.in/wishlist/view/da0c1e/",
}

// The main function sets up a web server with various handlers and options, including a background process for saving data to a database periodically.
func main() {
	flag.StringVar(&host, "host", "localhost", "Hostname")
	flag.StringVar(&port, "port", "8888", "Port number")
	web_only := flag.Bool("web", false, "Run only web server")
	flag.Parse()

	if !*web_only {
		go func() {
			for {
				parser.SaveToDB(wishlist_urls)
				time.Sleep(time.Hour)
			}
		}()
	}

	server := &http.Server{Addr: ":" + port}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/products", productsHandler)
	http.HandleFunc("/api/prices", pricesHandler)
	http.HandleFunc("/api/reset", resetHandler)
	http.HandleFunc("/api/new", newDataHandler)

	shutdownChan := make(chan bool, 1)
	go func() {
		log.Println("Server running on http://" + host + ":" + port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		shutdownChan <- true
	}()

	handleShutdown(server)
	<-shutdownChan
}
