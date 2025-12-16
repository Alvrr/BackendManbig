package repository

import (
	"backend/config"
	"backend/models"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func stokCol() *mongo.Collection { return config.DB.Collection("stok") }

func EnsureStokIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Index produk_id + created_at
	_, err := stokCol().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "produk_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

func CreateMutasi(m *models.StokMutasi) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Validate produk exists
	var tmp struct {
		ID string `bson:"_id"`
	}
	if err := config.ProdukCollection.FindOne(ctx, bson.M{"_id": m.ProdukID}).Decode(&tmp); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("produk tidak ditemukan")
		}
		return nil, err
	}
	return stokCol().InsertOne(ctx, m)
}

func GetMutasiByProduk(produkID string, page, pageSize int) ([]models.StokMutasi, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	}
	cur, err := stokCol().Find(ctx, bson.M{"produk_id": produkID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []models.StokMutasi
	for cur.Next(ctx) {
		var m models.StokMutasi
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func GetSaldoProduk(produkID string) (*models.StokSaldo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{Key: "produk_id", Value: produkID}}}},
		{{"$group", bson.D{
			{Key: "_id", Value: "$produk_id"},
			{Key: "masuk", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{bson.D{{Key: "$eq", Value: bson.A{"$jenis", "masuk"}}}, "$jumlah", 0}}}}}},
			{Key: "keluar", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{bson.D{{Key: "$eq", Value: bson.A{"$jenis", "keluar"}}}, "$jumlah", 0}}}}}},
		}}},
		{{"$project", bson.D{
			{Key: "produk_id", Value: "$_id"},
			{Key: "masuk", Value: 1},
			{Key: "keluar", Value: 1},
			{Key: "saldo", Value: bson.D{{Key: "$subtract", Value: bson.A{"$masuk", "$keluar"}}}},
		}}},
	}
	cur, err := stokCol().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Next(ctx) {
		var s models.StokSaldo
		if err := cur.Decode(&s); err != nil {
			return nil, err
		}
		return &s, nil
	}
	return &models.StokSaldo{ProdukID: produkID, Masuk: 0, Keluar: 0, Saldo: 0}, nil
}
