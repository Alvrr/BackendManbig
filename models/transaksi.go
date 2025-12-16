package models

import "time"

type TransaksiItem struct {
	ProdukID   string  `json:"produk_id" bson:"produk_id"`
	NamaProduk string  `json:"nama_produk,omitempty" bson:"nama_produk,omitempty"`
	Jumlah     int     `json:"jumlah" bson:"jumlah"`
	Harga      float64 `json:"harga" bson:"harga"`
}

type Transaksi struct {
	ID          string          `json:"id" bson:"_id"`
	KasirID     string          `json:"kasir_id" bson:"kasir_id"`
	PelangganID string          `json:"pelanggan_id" bson:"pelanggan_id"`
	TotalProduk int             `json:"total_produk" bson:"total_produk"`
	TotalHarga  float64         `json:"total_harga" bson:"total_harga"`
	Status      string          `json:"status" bson:"status"`
	Items       []TransaksiItem `json:"items,omitempty" bson:"items,omitempty"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
}
