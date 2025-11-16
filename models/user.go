package models

import "time"

type User struct {
	ID        string    `json:"id" bson:"_id"`
	Nama      string    `json:"nama" bson:"nama"`
	Email     string    `json:"email" bson:"email"`
	Password  string    `json:"password,omitempty" bson:"password"`
	Role      string    `json:"role" bson:"role"`
	NoHP      string    `json:"no_hp,omitempty" bson:"no_hp,omitempty"`
	Alamat    string    `json:"alamat,omitempty" bson:"alamat,omitempty"`
	Status    string    `json:"status" bson:"status"` // aktif/nonaktif
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
