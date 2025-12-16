package models

import (
	"time"
)

type Produk struct {
	ID         string    `json:"id" bson:"_id"`
	NamaProduk string    `json:"nama_produk" bson:"nama_produk"`
	KategoriID string    `json:"kategori_id" bson:"kategori_id"`
	Deskripsi  string    `json:"deskripsi" bson:"deskripsi"`
	HargaBeli  float64   `json:"harga_beli" bson:"harga_beli"`
	HargaJual  float64   `json:"harga_jual" bson:"harga_jual"`
	Stok       int       `json:"stok" bson:"stok"`
	Aktif      bool      `json:"aktif" bson:"aktif"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

// ProdukSwagger adalah struct khusus untuk dokumentasi Swagger response
// Menggunakan time.Time yang dikenal oleh Swagger instead of primitive.DateTime
type ProdukSwagger struct {
	ID         string    `json:"id" example:"PRD001"`
	NamaProduk string    `json:"nama_produk" example:"Beras 5kg"`
	KategoriID string    `json:"kategori_id" example:"KTG001"`
	Deskripsi  string    `json:"deskripsi" example:"Beras premium wangi pandan"`
	HargaBeli  float64   `json:"harga_beli" example:"60000"`
	HargaJual  float64   `json:"harga_jual" example:"65000"`
	Stok       int       `json:"stok" example:"100"`
	Aktif      bool      `json:"aktif" example:"true"`
	CreatedAt  time.Time `json:"created_at" example:"2025-01-01T10:00:00Z"`
}

// ProdukInput adalah struct untuk input data produk (tanpa ID dan CreatedAt)
type ProdukInput struct {
	NamaProduk string  `json:"nama_produk" example:"Beras 5kg"`
	KategoriID string  `json:"kategori_id" example:"KTG001"`
	Deskripsi  string  `json:"deskripsi" example:"Beras premium wangi pandan"`
	HargaBeli  float64 `json:"harga_beli" example:"60000"`
	HargaJual  float64 `json:"harga_jual" example:"65000"`
	Stok       int     `json:"stok" example:"100"`
	Aktif      bool    `json:"aktif" example:"true"`
}
