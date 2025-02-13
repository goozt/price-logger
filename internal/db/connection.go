package db

import (
	"context"
	"dilogger/internal/utils"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

var mu sync.Mutex

// The Product struct defines the attributes of a product including name, stock, price, creation date, and last update date.
type Product struct {
	Id        uint64    `json:"id"`
	Name      string    `json:"name"`
	Stock     int32     `json:"stock"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// The `CheckIfEquals` method is a function that compares the attributes of two `Product` instances to determine if they are equal.
func (p *Product) CheckIfEquals(e Product) bool {
	return p.Name == e.Name && p.Stock == e.Stock && p.Price == e.Price
}

// The `Connection` type represents a database connection in Go with associated context, driver connection, and database name.
type Connection struct {
	ctx    context.Context
	conn   driver.Conn
	dbname string
}

// The function `getOptions` returns a pointer to a `clickhouse.Options` struct with environment-specific configuration values.
func getOptions() *clickhouse.Options {
	env := utils.GetDBEnvironment()
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

func GetID() uint64 {
	mu.Lock()
	defer mu.Unlock()
	var maxID uint64
	c := ConnectDB("autoIncrement")
	err := c.conn.QueryRow(context.Background(), "SELECT ifNull(max(Id), 0) FROM autoIncrement").Scan(&maxID)
	if err != nil {
		log.Fatal(err)
	}
	return maxID + 1
}

func SetID(newID uint64) {
	c := ConnectDB("autoIncrement")
	err := c.conn.Exec(context.Background(), fmt.Sprintf("INSERT INTO autoIncrement VALUES (%d)", newID))
	if err != nil {
		log.Fatal(err)
	}
}

// This `CreateTable` method is responsible for creating a table in the ClickHouse database if it does not already exist.
func (c *Connection) CreateTable() {
	err := c.conn.Exec(context.Background(), `
	    CREATE TABLE IF NOT EXISTS `+c.dbname+` (
			Id UInt64,
			Name String,
			Stock Int32,
			Price Float64,
	        CreatedAt DateTime('Asia/Kolkata'),
	        UpdatedAt DateTime('Asia/Kolkata')
	    )
		ENGINE = MergeTree
		PRIMARY KEY (Name, Stock, Price)
		ORDER BY (Name, Stock, Price, DATE(UpdatedAt))
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
		"SELECT * FROM "+c.dbname,
	); err != nil {
		log.Fatal(err)
	}
	return
}

// The `GetUniqueNameList` method is responsible for fetching a list of unique product names from the ClickHouse database.
func (c *Connection) GetUniqueNameList() []string {
	var products []Product
	if err := c.conn.Select(
		context.Background(),
		&products,
		"SELECT DISTINCT Name FROM "+c.dbname+" ORDER BY Name ASC",
	); err != nil {
		log.Println(err)
	}

	var names []string
	for _, p := range products {
		names = append(names, p.Name)
	}
	return names
}

// The `GetListByName` method is responsible for fetching a list of `Product` structs from the ClickHouse database based on a specified product name.
func (c *Connection) GetListByName(name string) (products []Product) {
	if err := c.conn.Select(
		context.Background(),
		&products,
		"SELECT * FROM "+c.dbname+" WHERE Name='"+name+"'",
	); err != nil {
		log.Fatal(err)
	}
	return
}

// The `GetRecentList` method  is responsible for fetching a list of `Product` structs from the ClickHouse database where the record is created within a specified number of days.
func (c *Connection) GetRecentList(days int) (products []Product) {
	if err := c.conn.Select(
		context.Background(),
		&products,
		fmt.Sprintf("SELECT * FROM %s WHERE toDate(CreatedAt) = today()-INTERVAL %d DAY ORDER BY UpdatedAt", c.dbname, days),
	); err != nil {
		log.Fatal(err)
	}
	return
}

// The `GetRangeList` method  is responsible for fetching a list of `Product` structs from the ClickHouse database within a specified date range.
func (c *Connection) GetRangeList(startDate time.Time, endDate time.Time) (products []Product) {
	if err := c.conn.Select(
		context.Background(),
		&products,
		fmt.Sprintf("SELECT * FROM %s WHERE toDate(UpdatedAt) <= toDate(%s) AND toDate(CreatedAt) >= toDate(%s) ORDER BY UpdatedAt", c.dbname, startDate, endDate),
	); err != nil {
		log.Fatal(err)
	}
	return
}

// The `DeleteFromLast` method  is responsible for deleting a specified number of rows from the table in the ClickHouse database.
func (c *Connection) DeleteFromLast(limit int) {
	if err := c.conn.Exec(context.Background(), fmt.Sprintf(`DELETE FROM `+c.dbname+` ORDER BY UpdatedAt DESC LIMIT %d
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
	/*
		DROP TABLE IF EXISTS autoIncrement;
		CREATE TABLE IF NOT EXISTS autoIncrement ( Id UInt64 ) ENGINE=MergeTree ORDER BY Id;
		INSERT INTO autoIncrement VALUES (0);
		SELECT * FROM autoIncrement;
	*/
	c.RemoveList()
	c.CreateTable()
}

// This function `InsertOne` is responsible for inserting a single `Product` struct into the ClickHouse database using a batch operation.
func (c *Connection) InsertOne(product Product) {
	batch, err := c.conn.PrepareBatch(c.ctx, "INSERT INTO "+c.dbname)
	if err != nil {
		log.Fatal(err)
	}
	product.Id = GetID()
	if err := batch.AppendStruct(&product); err != nil {
		log.Fatal(err)
	}
	if err := batch.Send(); err != nil {
		log.Fatal(err)
	}
	SetID(product.Id)
}

func (c *Connection) InsertMultiple(products []Product) {
	for _, product := range products {
		if c.UpdateRecord(product) {
			println("Updated")
		} else {
			newId := GetID()
			now := time.Now()
			if err := c.conn.AsyncInsert(
				c.ctx,
				fmt.Sprintf(
					`INSERT INTO %s VALUES (%d,'%s',%d,%0.2f,toDateTime('%s'),toDateTime('%s'))`,
					c.dbname,
					newId,
					product.Name,
					product.Stock,
					product.Price,
					now,
					now,
				),
				true,
			); err != nil {
				log.Fatal(err)
			}
			SetID(newId)
		}
	}
}

func (c *Connection) UpdateRecord(new_product Product) bool {
	query := "SELECT * FROM %s WHERE Name = '%s' AND Stock = %d AND Price = %0.2f"
	var old_product Product
	if err := c.conn.QueryRow(context.Background(), fmt.Sprintf(query, c.dbname, new_product.Name, new_product.Stock, new_product.Price)).ScanStruct(&old_product); err != nil {
		return false
	}
	if !new_product.CheckIfEquals(old_product) {
		return false
	}
	query = `ALTER TABLE %s UPDATE Stock = %d, Price = %f, UpdatedAt = toDateTime('%s') WHERE Id = %d`

	if err := c.conn.Exec(
		context.Background(),
		fmt.Sprintf(
			query,
			c.dbname,
			new_product.Stock,
			new_product.Price,
			new_product.CreatedAt.Format(time.DateTime),
			old_product.Id,
		),
	); err != nil {
		log.Println("INSERT ERROR: ", err)
		return false
	}
	return true
}
