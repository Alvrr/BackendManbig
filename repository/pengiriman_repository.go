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

func pengirimanCol() *mongo.Collection { return config.PengirimanCollection }

func EnsurePengirimanIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := pengirimanCol().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "driver_id", Value: 1}}},
		{Keys: bson.D{{Key: "transaksi_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

func CreatePengiriman(p models.Pengiriman) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return pengirimanCol().InsertOne(ctx, p)
}

func GetPengirimanFiltered(filter bson.M) ([]models.Pengiriman, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := pengirimanCol().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []models.Pengiriman
	for cur.Next(ctx) {
		var p models.Pengiriman
		if err := cur.Decode(&p); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func GetPengirimanByID(id string) (*models.Pengiriman, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var p models.Pengiriman
	if err := pengirimanCol().FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func UpdatePengiriman(id string, p models.Pengiriman) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	set := bson.M{}
	if p.Jenis != "" {
		set["jenis"] = p.Jenis
	}
	if p.Ongkir != 0 {
		set["ongkir"] = p.Ongkir
	}
	if p.Status != "" {
		set["status"] = p.Status
	}
	upd := bson.M{"$set": set}
	return pengirimanCol().UpdateOne(ctx, bson.M{"_id": id}, upd)
}

func DeletePengiriman(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return pengirimanCol().DeleteOne(ctx, bson.M{"_id": id})
}
