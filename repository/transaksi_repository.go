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

func transaksiCol() *mongo.Collection { return config.DB.Collection("transaksi") }

func EnsureTransaksiIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := transaksiCol().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "kasir_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

func ListTransaksi(filter bson.M, page, pageSize int) ([]models.Transaksi, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if filter == nil {
		filter = bson.M{}
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize)).SetLimit(int64(pageSize))
	}
	cur, err := transaksiCol().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []models.Transaksi
	for cur.Next(ctx) {
		var t models.Transaksi
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func GetTransaksiByID(id string) (*models.Transaksi, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var t models.Transaksi
	if err := transaksiCol().FindOne(ctx, bson.M{"_id": id}).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func CreateTransaksi(t *models.Transaksi) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return transaksiCol().InsertOne(ctx, t)
}

func UpdateTransaksi(id string, update bson.M) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return transaksiCol().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
}

func DeleteTransaksi(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return transaksiCol().DeleteOne(ctx, bson.M{"_id": id})
}

// BestSellerItem represents aggregation result for best sellers
type BestSellerItem struct {
	ProdukID string `bson:"_id" json:"produk_id"`
	Nama     string `bson:"nama" json:"nama"`
	Jumlah   int    `bson:"jumlah" json:"jumlah"`
}

// GetBestSellers aggregates transaksi items within date range and returns top products
func GetBestSellers(start, end time.Time, limit int) ([]BestSellerItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	matchStage := bson.D{{Key: "$match", Value: bson.M{"created_at": bson.M{"$gte": start, "$lte": end}}}}
	unwindStage := bson.D{{Key: "$unwind", Value: "$items"}}
	groupStage := bson.D{{Key: "$group", Value: bson.M{
		"_id":    "$items.produk_id",
		"jumlah": bson.M{"$sum": "$items.jumlah"},
	}}}
	sortStage := bson.D{{Key: "$sort", Value: bson.M{"jumlah": -1}}}
	limitStage := bson.D{{Key: "$limit", Value: limit}}
	lookupStage := bson.D{{Key: "$lookup", Value: bson.M{
		"from":         "produk",
		"localField":   "_id",
		"foreignField": "_id",
		"as":           "produk",
	}}}
	addFields := bson.D{{Key: "$addFields", Value: bson.M{
		"nama": bson.M{"$ifNull": bson.A{bson.M{"$arrayElemAt": bson.A{"$produk.nama_produk", 0}}, "$_id"}},
	}}}
	projectStage := bson.D{{Key: "$project", Value: bson.M{"produk": 0}}}

	pipeline := mongo.Pipeline{matchStage, unwindStage, groupStage, sortStage, limitStage, lookupStage, addFields, projectStage}
	cur, err := transaksiCol().Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []BestSellerItem
	for cur.Next(ctx) {
		var it BestSellerItem
		if err := cur.Decode(&it); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, nil
}
