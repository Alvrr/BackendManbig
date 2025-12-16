package repository

import (
	"backend/config"
	"backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func pembayaranCol() *mongo.Collection { return config.PembayaranCollection }

func EnsurePembayaranIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := pembayaranCol().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "kasir_id", Value: 1}}},
		{Keys: bson.D{{Key: "transaksi_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

// Ambil semua pembayaran (admin/kasir)
func GetAllPembayaran() ([]models.Pembayaran, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := pembayaranCol().Find(ctx, bson.M{}, opts)
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

// Ambil pembayaran dengan filter (khusus driver)
func GetPembayaranFiltered(filter bson.M) ([]models.Pembayaran, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := pembayaranCol().Find(ctx, filter, opts)
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

// Ambil satu pembayaran berdasarkan ID
func GetPembayaranByID(id string) (*models.Pembayaran, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var pembayaran models.Pembayaran
	err := pembayaranCol().FindOne(ctx, bson.M{"_id": id}).Decode(&pembayaran)
	if err != nil {
		return nil, err
	}
	return &pembayaran, nil
}

// Tambah data pembayaran baru
func CreatePembayaran(p models.Pembayaran) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pembayaranCol().InsertOne(ctx, p)
}

// Update pembayaran (admin/kasir)
func UpdatePembayaran(id string, p models.Pembayaran) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	set := bson.M{}
	if p.Metode != "" {
		set["metode"] = p.Metode
	}
	if p.TotalBayar != 0 {
		set["total_bayar"] = p.TotalBayar
	}
	if p.Status != "" {
		set["status"] = p.Status
	}
	update := bson.M{"$set": set}
	return pembayaranCol().UpdateOne(ctx, bson.M{"_id": id}, update)
}

// Hapus pembayaran
func DeletePembayaran(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pembayaranCol().DeleteOne(ctx, bson.M{"_id": id})
}

// Set status transaksi jadi "Selesai"
func SelesaikanPembayaran(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": "Selesai"}}

	_, err := pembayaranCol().UpdateOne(ctx, filter, update)
	return err
}
