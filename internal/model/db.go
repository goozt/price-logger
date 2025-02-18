package model

import (
	"time"
)

type Product struct {
	Id        string    `form:"id" json:"id"`
	Name      string    `form:"name" json:"name"`
	Stock     int32     `form:"stock" json:"stock"`
	Price     float64   `form:"price" json:"price"`
	CreatedAt time.Time `form:"created" json:"created"`
	UpdatedAt time.Time `form:"updated" json:"updated"`
}
