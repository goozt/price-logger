package db

import (
	"context"
	"dilogger/utils"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Product struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Stock     int32     `json:"stock"`
	Price     float64   `json:"price"`
}

type Connection struct {
	ctx    context.Context
	conn   driver.Conn
	dbname string
}

func getOptions() *clickhouse.Options {
	env := utils.GetEnvironment()
	return &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", env.Host, env.Port)},
		Auth: clickhouse.Auth{
			Database: env.Database,
			Username: env.Username,
			Password: env.Password,
		},
	}
}

func ConnectDB(dbname string) Connection {
	ctx := context.Background()
	conn, err := clickhouse.Open(getOptions())

	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		log.Fatal(err)
	}
	return Connection{ctx, conn, dbname}
}

func (c *Connection) CreateTable() {
	err := c.conn.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS `+c.dbname+` (
	        Timestamp DateTime('Asia/Kolkata'),
			Name String,
			Stock Int32,
			Price Float64
	    )
		ENGINE = MergeTree ORDER BY (DATE(Timestamp), Name)
		TTL Timestamp + toIntervalDay(1) GROUP BY DATE(Timestamp) SET Price = min(Price)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) GetList() (products []Product) {
	if err := c.conn.Select(
		context.Background(),
		&products,
		"SELECT Timestamp,Name,Stock,Price FROM "+c.dbname,
	); err != nil {
		log.Fatal(err)
	}
	return
}

func (c *Connection) InsertOne(product Product) {
	batch, err := c.conn.PrepareBatch(c.ctx, "INSERT INTO "+c.dbname)
	if err != nil {
		log.Fatal(err)
	}
	if err := batch.AppendStruct(&product); err != nil {
		log.Fatal(err)
	}
	if err := batch.Send(); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) InsertMultiple(products []Product) {
	batch, err := c.conn.PrepareBatch(c.ctx, "INSERT INTO "+c.dbname)
	if err != nil {
		log.Fatal(err)
	}
	for _, product := range products {
		if err := batch.AppendStruct(&product); err != nil {
			log.Fatal(err)
		}
	}
	if err := batch.Send(); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) DeleteFromLast(limit int) {
	if err := c.conn.Exec(context.Background(), fmt.Sprintf(`DELETE FROM `+c.dbname+` ORDER BY Timestamp DESC limit %d
	`, limit)); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) RemoveList() {
	if err := c.conn.Exec(context.Background(), `DROP TABLE IF EXISTS `+c.dbname); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) CompactList() {
	if err := c.conn.Exec(context.Background(), `OPTIMIZE TABLE `+c.dbname); err != nil {
		log.Fatal(err)
	}
}

func (c *Connection) ResetDB() {
	c.RemoveList()
	c.CreateTable()
}
