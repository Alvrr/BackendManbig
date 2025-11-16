package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Produk struct {
	ID         string             `json:"id" bson:"_id"`
	NamaProduk string             `json:"nama_produk" bson:"nama_produk" validate:"required"`
	Kategori   string             `json:"kategori" bson:"kategori" validate:"required"`
	Harga      int                `json:"harga" bson:"harga" validate:"required"`
	Stok       int                `json:"stok" bson:"stok" validate:"required"`
	Deskripsi  string             `json:"deskripsi" bson:"deskripsi"`
	CreatedAt  primitive.DateTime `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

// ProdukSwagger adalah struct khusus untuk dokumentasi Swagger response
// Menggunakan time.Time yang dikenal oleh Swagger instead of primitive.DateTime
type ProdukSwagger struct {
	ID         string    `json:"id" example:"1"`
	NamaProduk string    `json:"nama_produk" example:"Galon Air Mineral"`
	Kategori   string    `json:"kategori" example:"Minuman"`
	Harga      int       `json:"harga" example:"20000"`
	Stok       int       `json:"stok" example:"100"`
	Deskripsi  string    `json:"deskripsi" example:"Air mineral kemasan galon 19 liter"`
	CreatedAt  time.Time `json:"created_at,omitempty" example:"2024-01-15T10:30:00Z"`
}

// ProdukInput adalah struct untuk input data produk (tanpa ID dan CreatedAt)
type ProdukInput struct {
	NamaProduk string `json:"nama_produk" example:"Galon Air Mineral" validate:"required"`
	Kategori   string `json:"kategori" example:"Minuman" validate:"required"`
	Harga      int    `json:"harga" example:"20000" validate:"required"`
	Stok       int    `json:"stok" example:"100" validate:"required"`
	Deskripsi  string `json:"deskripsi" example:"Air mineral kemasan galon 19 liter"`
}
