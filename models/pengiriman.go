package models

import "time"

type Pengiriman struct {
	ID          string    `json:"id" bson:"_id"`
	TransaksiID string    `json:"transaksi_id" bson:"transaksi_id"`
	DriverID    string    `json:"driver_id" bson:"driver_id"`
	Jenis       string    `json:"jenis" bson:"jenis"`
	Ongkir      float64   `json:"ongkir" bson:"ongkir"`
	Status      string    `json:"status" bson:"status"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}
