package repository

import (
	"context"
	"fmt"
	"time"

	"backend/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitializeCounters menginisialisasi counter yang diperlukan jika belum ada
func InitializeCounters() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	counters := []map[string]interface{}{
		{"_id": "admin", "prefix": "ADM", "sequence_value": 0},
		{"_id": "kasir", "prefix": "KSR", "sequence_value": 0},
		{"_id": "driver", "prefix": "DRV", "sequence_value": 0},
		{"_id": "pelanggan", "prefix": "PLG", "sequence_value": 0},
		{"_id": "produk", "prefix": "PRD", "sequence_value": 0},
		{"_id": "pembayaran", "prefix": "PMB", "sequence_value": 0},
	}

	for _, counter := range counters {
		filter := bson.M{"_id": counter["_id"]}
		update := bson.M{"$setOnInsert": counter}
		opts := options.Update().SetUpsert(true)

		_, err := config.CounterCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("gagal inisialisasi counter %s: %v", counter["_id"], err)
		}
	}

	return nil
}

// Fungsi generate ID dengan prefix dan angka urut 3 digit
func GenerateID(counterID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update sekaligus ambil nilai sequence_value terbaru (increment 1)
	filter := bson.M{"_id": counterID}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result struct {
		SequenceValue int    `bson:"sequence_value"`
		Prefix        string `bson:"prefix"`
	}

	err := config.CounterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("counter %s tidak ditemukan", counterID)
		}
		return "", err
	}

	// Format ID: prefix + 3 digit angka, contoh: P001, C007, TSX015
	newID := fmt.Sprintf("%s%03d", result.Prefix, result.SequenceValue)
	return newID, nil
}
func GenerateUserID(role string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mapping role ke _id counter (admin, kasir, driver)
	counterID := ""
	switch role {
	case "admin":
		counterID = "admin"
	case "kasir":
		counterID = "kasir"
	case "driver":
		counterID = "driver"
	default:
		return "", fmt.Errorf("role tidak dikenali: %s", role)
	}

	filter := bson.M{"_id": counterID}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var result struct {
		SequenceValue int    `bson:"sequence_value"`
		Prefix        string `bson:"prefix"`
	}

	err := config.CounterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("counter %s tidak ditemukan", counterID)
		}
		return "", err
	}

	newID := fmt.Sprintf("%s%03d", result.Prefix, result.SequenceValue)
	return newID, nil
}
