package repository

import (
	"backend/config"
	"backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GANTI yang ini:
// var produkCol *mongo.Collection = config.ProdukCollection

// JADI fungsi lazy load:
func produkCol() *mongo.Collection {
	return config.ProdukCollection
}

func GetAllProduk() ([]models.Produk, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := produkCol().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var produks []models.Produk
	for cursor.Next(ctx) {
		var p models.Produk
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		produks = append(produks, p)
	}

	return produks, nil
}

func GetProdukByID(id string) (*models.Produk, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var produk models.Produk
	err := produkCol().FindOne(ctx, bson.M{"_id": id}).Decode(&produk)
	if err != nil {
		return nil, err
	}
	return &produk, nil
}

func CreateProduk(p models.Produk) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return produkCol().InsertOne(ctx, p)
}

func UpdateProduk(id string, p models.Produk) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"nama_produk": p.NamaProduk,
			"kategori":    p.Kategori,
			"harga":       p.Harga,
			"stok":        p.Stok,
			"deskripsi":   p.Deskripsi,
		},
	}

	return produkCol().UpdateOne(ctx, bson.M{"_id": id}, update)
}

func DeleteProduk(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return produkCol().DeleteOne(ctx, bson.M{"_id": id})
}
