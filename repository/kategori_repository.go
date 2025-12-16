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

func kategoriCol() *mongo.Collection { return config.KategoriCollection }

func EnsureKategoriIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Unique index on nama_kategori
	model := mongo.IndexModel{
		Keys:    bson.D{{Key: "nama_kategori", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := kategoriCol().Indexes().CreateOne(ctx, model)
	return err
}

func GetAllKategori() ([]models.Kategori, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cur, err := kategoriCol().Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var list []models.Kategori
	for cur.Next(ctx) {
		var k models.Kategori
		if err := cur.Decode(&k); err != nil {
			return nil, err
		}
		list = append(list, k)
	}
	return list, nil
}

func GetKategoriByID(id string) (*models.Kategori, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var k models.Kategori
	if err := kategoriCol().FindOne(ctx, bson.M{"_id": id}).Decode(&k); err != nil {
		return nil, err
	}
	return &k, nil
}

func CreateKategori(k *models.Kategori) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return kategoriCol().InsertOne(ctx, k)
}

func UpdateKategori(id string, update bson.M) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return kategoriCol().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
}

func DeleteKategori(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return kategoriCol().DeleteOne(ctx, bson.M{"_id": id})
}
