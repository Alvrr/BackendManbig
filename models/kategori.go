package models

import "time"

type Kategori struct {
	ID           string    `json:"id" bson:"_id"`
	NamaKategori string    `json:"nama_kategori" bson:"nama_kategori"`
	Deskripsi    string    `json:"deskripsi,omitempty" bson:"deskripsi,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}
