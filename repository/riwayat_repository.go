package repository

import (
	"backend/config"
	"backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Ambil pembayaran yang sudah selesai (riwayat)
func GetRiwayatPembayaran(filter bson.M) ([]models.Pembayaran, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter["status"] = "Selesai"
	cursor, err := config.PembayaranCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var list []models.Pembayaran
	for cursor.Next(ctx) {
		var p models.Pembayaran
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}
