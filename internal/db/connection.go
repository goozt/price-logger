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

// The `Product` type represents a product with timestamp, name, stock quantity, and price fields.
type Product struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	Stock     int32     `json:"stock"`
	Price     float64   `json:"price"`
}

// The `Connection` type represents a database connection in Go with associated context, driver connection, and database name.
type Connection struct {
	ctx    context.Context
	conn   driver.Conn
	dbname string
}

// The function `getOptions` returns a pointer to a `clickhouse.Options` struct with environment-specific configuration values.
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

// The ConnectDB function establishes a connection to a ClickHouse database and returns a Connection object.
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

// This `CreateTable` method is responsible for creating a table in the ClickHouse database if it does not already exist.
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

// The `GetList` method is responsible for fetching a list of `Product` structs from the ClickHouse database.
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

// This function `InsertOne` is responsible for inserting a single `Product` struct into the ClickHouse database using a batch operation.
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

// This function `InsertMultiple` in the `Connection` struct is responsible for inserting multiple `Product` structs into the ClickHouse database in a batch operation.
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

// The `DeleteFromLast` method  is responsible for deleting a specified number of rows from the table in the ClickHouse database.
func (c *Connection) DeleteFromLast(limit int) {
	if err := c.conn.Exec(context.Background(), fmt.Sprintf(`DELETE FROM `+c.dbname+` ORDER BY Timestamp DESC limit %d
	`, limit)); err != nil {
		log.Fatal(err)
	}
}

// The `RemoveList` method is responsible for deleting the table inside database if it exists.
func (c *Connection) RemoveList() {
	if err := c.conn.Exec(context.Background(), `DROP TABLE IF EXISTS `+c.dbname); err != nil {
		log.Fatal(err)
	}
}

// The `CompactList` method is responsible for optimizing the table in the ClickHouse database.
func (c *Connection) CompactList() {
	if err := c.conn.Exec(context.Background(), `OPTIMIZE TABLE `+c.dbname); err != nil {
		log.Fatal(err)
	}
}

// The `ResetDB` method in the `Connection` struct is a function that resets the database by first removing the existing table and then creating a new table with the same name and structure.
func (c *Connection) ResetDB() {
	c.RemoveList()
	c.CreateTable()
}
