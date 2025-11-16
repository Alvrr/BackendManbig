package models

type Pelanggan struct {
	ID     string `json:"id" bson:"_id"`
	Nama   string `json:"nama" bson:"nama" validate:"required"`
	Email  string `json:"email" bson:"email" validate:"required,email"`
	NoHP   string `json:"no_hp" bson:"no_hp" validate:"required"`
	Alamat string `json:"alamat" bson:"alamat" validate:"required"`
}
