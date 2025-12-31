package models

import "time"

type StokMutasi struct {
	ID         string    `json:"id" bson:"_id"`
	ProdukID   string    `json:"produk_id" bson:"produk_id"`
	Jenis      string    `json:"jenis" bson:"jenis"` // masuk / keluar / adjust
	Jumlah     int       `json:"jumlah" bson:"jumlah"`
	UserID     string    `json:"user_id" bson:"user_id"`
	RefID      string    `json:"ref_id,omitempty" bson:"ref_id,omitempty"`
	RefType    string    `json:"ref_type,omitempty" bson:"ref_type,omitempty"` // contoh: transaksi, manual
	Keterangan string    `json:"keterangan,omitempty" bson:"keterangan,omitempty"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
}

type StokSaldo struct {
	ProdukID string `json:"produk_id" bson:"produk_id"`
	Masuk    int    `json:"masuk" bson:"masuk"`
	Keluar   int    `json:"keluar" bson:"keluar"`
	Saldo    int    `json:"saldo" bson:"saldo"`
}
