package db

import (
	"dilogger/internal/model"
	"dilogger/internal/push"
	"log/slog"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/hook"
	"github.com/spf13/cobra"
)

type Server struct {
	App               *pocketbase.PocketBase
	productCollection *core.Collection
	priceCollection   *core.Collection
	urlCollection     *core.Collection
	Notification      *push.OneSignalApp
	logger            *slog.Logger
}

// Cobra's AddCommand function extended
func (s *Server) AddCobraCommand(cmds ...*cobra.Command) {
	s.App.RootCmd.AddCommand(cmds...)
}

// Pocketbase OnServe hook extended
func (s *Server) OnServe() *hook.Hook[*core.ServeEvent] {
	return s.App.OnServe()
}

// Pocketbase Logger extended
func (s *Server) Logger() *slog.Logger {
	return s.logger
}

// Pocketbase Cron function extended
func (s *Server) Cron() *cron.Cron {
	return s.App.Cron()
}

// Create new Url collection in database
func (s *Server) NewUrlCollection() {
	collection, err := s.App.FindCollectionByNameOrId("urls")
	if err == nil {
		s.urlCollection = collection
		return
	}
	collection = NewCollection("urls")
	err = s.App.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.urlCollection = collection
}

// Create new Product collection in database
func (s *Server) NewProductCollection() {
	if s.urlCollection == nil {
		s.NewUrlCollection()
	}
	collection, err := s.App.FindCollectionByNameOrId("products")
	if err == nil {
		s.productCollection = collection
		return
	}
	collection = NewCollection("products")
	err = s.App.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.productCollection = collection
}

// Create new Price Collection in database
func (s *Server) NewPriceCollection() {
	if s.productCollection == nil {
		s.NewProductCollection()
	}
	collection, err := s.App.FindCollectionByNameOrId("prices")
	if err == nil {
		s.priceCollection = collection
		return
	}
	collection = NewCollection("prices", s.productCollection.Id)
	err = s.App.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.priceCollection = collection
}

// Get list of URLs from url database
func (s *Server) GetURLs() []string {
	var urls []string
	records, err := s.App.FindAllRecords(s.urlCollection)
	if err != nil {
		s.logger.Error(err.Error())
		return urls
	}
	for _, record := range records {
		urls = append(urls, record.GetString("url"))
	}
	return urls
}

// Check for matching products with similar pricing in the database
func (s *Server) PriceMatch(product model.Product) (*core.Record, *core.Record, bool) {
	productRecord, err := s.App.FindFirstRecordByData("products", "name", product.Name)
	if err != nil {
		return nil, nil, false
	}
	records, err := s.App.FindRecordsByFilter(
		s.priceCollection.Name,
		"product = {:product} && price = {:price}",
		"-updated", 1, 0,
		dbx.Params{
			"product": productRecord.Id,
			"price":   product.Price,
		},
	)
	if err != nil || len(records) < 1 {
		return productRecord, nil, false
	}
	return productRecord, records[0], true
}

// Add product data to the database
func (s *Server) AddToCollection(products []model.Product) {
	if s.priceCollection == nil || s.productCollection == nil {
		s.NewPriceCollection()
	}
	for _, product := range products {
		productRecord, record, matched := s.PriceMatch(product)
		if matched {
			if stock := int32(productRecord.Get("stock").(float64)); stock != product.Stock {
				productRecord.Set("stock", product.Stock)
			}
			record.Set("price", product.Price)
		} else {
			if productRecord == nil {
				productRecord = core.NewRecord(s.productCollection)
				productRecord.Set("name", product.Name)
				productRecord.Set("stock", product.Stock)
				err := s.App.Save(productRecord)
				if err != nil {
					s.logger.Error(err.Error())
					return
				}
			}
			record = core.NewRecord(s.priceCollection)
			record.Set("product", productRecord.Id)
			record.Set("price", product.Price)
		}
		err := s.App.Save(record)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
	}
}

// Update price when the hook is triggered
func (s *Server) PriceUpdateHook(bindingFunction func(e *core.RecordEvent) error) {
	s.App.OnRecordCreate("prices").BindFunc(func(e *core.RecordEvent) error {
		price := e.Record.GetFloat("price")
		existingRecords, err := s.App.FindRecordsByFilter(
			"prices",
			"product = {:product} && price = {:price}",
			"-updated", 1, 0,
			dbx.Params{
				"product": e.Record.GetString("product"),
				"price":   price,
			},
		)
		if err != nil {
			return err
		}
		if len(existingRecords) > 0 {
			existingRecord := existingRecords[0]
			existingRecord.Set("price", price)
			if err := s.App.Save(existingRecord); err != nil {
				return err
			}
			return nil
		}
		return e.Next()
	})
	s.App.OnRecordAfterCreateSuccess("prices").BindFunc(bindingFunction)
}

// Create Product object from product record
func (s *Server) GetProduct(priceRecord *core.Record) model.Product {
	var product model.Product
	product.Id = priceRecord.Id
	product.Price = priceRecord.GetFloat("price")
	product.CreatedAt = priceRecord.GetDateTime("created").Time()
	product.UpdatedAt = priceRecord.GetDateTime("updated").Time()
	s.App.ExpandRecord(priceRecord, []string{"product"}, nil)
	record := priceRecord.ExpandedOne("product")
	if record == nil {
		return model.Product{}
	}
	product.Name = record.GetString("name")
	product.Stock = int32(record.GetInt("stock"))
	return product
}

// Count the number of records with same product id inside prices collection
func (s *Server) CountPriceRecords(productId string) int64 {
	count, err := s.App.CountRecords("prices", dbx.HashExp{"product": productId})
	if err != nil {
		s.logger.Error(err.Error())
		return 0
	}
	return count
}
