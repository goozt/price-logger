package product

import (
	"dilogger/internal/db"
	"dilogger/internal/model"
	"dilogger/internal/parser"
	"sync"
)

// The GetProducts function concurrently fetches and parses product data from multiple URLs using goroutines and channels.
func GetProducts(urls []string) (products []model.Product) {

	var wg sync.WaitGroup
	ch := make(chan model.Product)

	for _, url := range urls {
		wg.Add(1)
		go parser.Parse(ch, &wg, url)
	}

	// This block is creating an anonymous goroutine that reads from the channel `ch` and appends the received `model.Product` objects to the `products` slice.
	go func() {
		for product := range ch {
			products = append(products, product)
		}
	}()

	wg.Wait()
	close(ch)

	return
}

func ReloadData(server *db.Server) {
	urls := server.GetURLs()
	server.AddToCollection(GetProducts(urls))
}
