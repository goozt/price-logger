package monitor

import (
	"dilogger/internal/db"
	"dilogger/internal/push"
	"dilogger/internal/utils"
	"sync"
	"time"
)

var (
	conn       = db.ConnectDB(utils.GetEnv("DB_TABLE", "dilogger"))
	lastPrices = make(map[string]float64)
	mu         sync.Mutex
)

// The `MonitorPriceChanges` function continuously checks for price changes in products and sends notifications if a change is detected.
func MonitorPriceChanges() {
	for {
		products := conn.GetRecentList(1)

		mu.Lock()
		for _, product := range products {
			lastPrice, exists := lastPrices[product.Name]

			if !exists || lastPrice != product.Price {
				if exists {
					push.SendNotification(product)
				}
				lastPrices[product.Name] = product.Price
			}
		}
		mu.Unlock()

		time.Sleep(10 * time.Second)
	}
}
