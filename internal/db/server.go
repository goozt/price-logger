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
	app               *pocketbase.PocketBase
	productCollection *core.Collection
	priceCollection   *core.Collection
	urlCollection     *core.Collection
	Notification      *push.OneSignalApp
	logger            *slog.Logger
}

func (s *Server) Start() {
	if err := s.app.Start(); err != nil {
		s.logger.Error(err.Error())
		return
	}
}

func (s *Server) OnServe() *hook.Hook[*core.ServeEvent] {
	return s.app.OnServe()
}

func (s *Server) Logger() *slog.Logger {
	return s.logger
}

func (s *Server) Cron() *cron.Cron {
	return s.app.Cron()
}

func (s *Server) NewUrlCollection() {
	collection, err := s.app.FindCollectionByNameOrId("urls")
	if err == nil {
		s.urlCollection = collection
		return
	} else {
		s.logger.Error(err.Error())
	}
	collection = NewCollection("urls")
	err = s.app.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.urlCollection = collection
}

func (s *Server) NewProductCollection() {
	if s.urlCollection == nil {
		s.NewUrlCollection()
	}
	collection, err := s.app.FindCollectionByNameOrId("products")
	if err == nil {
		s.productCollection = collection
		return
	} else {
		s.logger.Error(err.Error())
	}
	collection = NewCollection("products")
	err = s.app.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.productCollection = collection
}

func (s *Server) NewPriceCollection() {
	if s.productCollection == nil {
		s.NewProductCollection()
	}
	collection, err := s.app.FindCollectionByNameOrId("prices")
	if err == nil {
		s.priceCollection = collection
		return
	} else {
		s.logger.Error(err.Error())
	}
	collection = NewCollection("prices", s.productCollection.Id)
	err = s.app.Save(collection)
	if err != nil {
		s.logger.Error(err.Error())
		return
	}
	s.priceCollection = collection
}

func (s *Server) GetURLs() []string {
	var urls []string
	records, err := s.app.FindAllRecords(s.urlCollection)
	if err != nil {
		s.logger.Error(err.Error())
		return urls
	}
	for _, record := range records {
		urls = append(urls, record.GetString("url"))
	}
	return urls
}

func (s *Server) PriceMatch(product model.Product) (*core.Record, *core.Record, bool) {
	productRecord, err := s.app.FindFirstRecordByData("products", "name", product.Name)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, nil, false
	}
	records, err := s.app.FindRecordsByFilter(
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
				err := s.app.Save(productRecord)
				if err != nil {
					s.logger.Error(err.Error())
					return
				}
			}
			record = core.NewRecord(s.priceCollection)
			record.Set("product", productRecord.Id)
			record.Set("price", product.Price)
		}
		err := s.app.Save(record)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
	}
}

func (s *Server) AddCommand(command string, Fn func(app core.App)) {
	s.app.RootCmd.AddCommand(&cobra.Command{
		Use: command,
		Run: func(cmd *cobra.Command, args []string) {
			Fn(s.app)
		},
	})
}

func (s *Server) PriceUpdateHook(bindingFunction func(e *core.RecordEvent) error) {
	s.app.OnRecordCreate("prices").BindFunc(func(e *core.RecordEvent) error {
		price := e.Record.GetFloat("price")
		existingRecords, err := s.app.FindRecordsByFilter(
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
			if err := s.app.Save(existingRecord); err != nil {
				return err
			}
			return nil
		}
		return e.Next()
	})
	s.app.OnRecordAfterCreateSuccess("prices").BindFunc(bindingFunction)
}

func (s *Server) GetProduct(priceRecord *core.Record) model.Product {
	var product model.Product
	product.Id = priceRecord.Id
	product.Price = priceRecord.GetFloat("price")
	product.CreatedAt = priceRecord.GetDateTime("created").Time()
	product.UpdatedAt = priceRecord.GetDateTime("updated").Time()
	s.app.ExpandRecord(priceRecord, []string{"product"}, nil)
	record := priceRecord.ExpandedOne("product")
	if record == nil {
		return model.Product{}
	}
	product.Name = record.GetString("name")
	product.Stock = int32(record.GetInt("stock"))
	return product
}

func (s *Server) CountPriceRecords(productId string) int64 {
	count, err := s.app.CountRecords("prices", dbx.HashExp{"product": productId})
	if err != nil {
		s.logger.Error(err.Error())
		return 0
	}
	return count
}
