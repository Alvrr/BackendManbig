package repository

import (
	"backend/config"
	"backend/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// lazy-load koleksi pelanggan
func pelangganCol() *mongo.Collection {
	return config.PelangganCollection
}

func GetAllPelanggan() ([]models.Pelanggan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := pelangganCol().Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []models.Pelanggan
	for cursor.Next(ctx) {
		var p models.Pelanggan
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		list = append(list, p)
	}

	return list, nil
}

func GetPelangganByID(id string) (*models.Pelanggan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var pelanggan models.Pelanggan
	err := pelangganCol().FindOne(ctx, bson.M{"_id": id}).Decode(&pelanggan)
	if err != nil {
		return nil, err
	}
	return &pelanggan, nil
}

func CreatePelanggan(p models.Pelanggan) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pelangganCol().InsertOne(ctx, p)
}

func UpdatePelanggan(id string, p models.Pelanggan) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"nama":   p.Nama,
			"email":  p.Email,
			"no_hp":  p.NoHP,
			"alamat": p.Alamat,
		},
	}

	return pelangganCol().UpdateOne(ctx, bson.M{"_id": id}, update)
}

func DeletePelanggan(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return pelangganCol().DeleteOne(ctx, bson.M{"_id": id})
}
