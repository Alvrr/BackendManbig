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

// GANTI yang ini:
// var produkCol *mongo.Collection = config.ProdukCollection

// JADI fungsi lazy load:
func produkCol() *mongo.Collection {
	return config.ProdukCollection
}

func EnsureProdukIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := produkCol().Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "kategori_id", Value: 1}}})
	return err
}

func GetAllProduk() ([]models.Produk, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := produkCol().Find(ctx, bson.M{}, opts)
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

	if p.KategoriID != "" {
		var tmp struct {
			ID string `bson:"_id"`
		}
		if err := config.KategoriCollection.FindOne(ctx, bson.M{"_id": p.KategoriID}).Decode(&tmp); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("kategori tidak ditemukan")
			}
			return nil, err
		}
	}

	return produkCol().InsertOne(ctx, p)
}

func UpdateProduk(id string, p models.Produk) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	set := bson.M{}
	if p.NamaProduk != "" {
		set["nama_produk"] = p.NamaProduk
	}
	if p.KategoriID != "" {
		set["kategori_id"] = p.KategoriID
	}
	if p.Deskripsi != "" {
		set["deskripsi"] = p.Deskripsi
	}
	if p.HargaBeli != 0 {
		set["harga_beli"] = p.HargaBeli
	}
	if p.HargaJual != 0 {
		set["harga_jual"] = p.HargaJual
	}
	if p.Stok != 0 {
		set["stok"] = p.Stok
	}
	set["aktif"] = p.Aktif
	// Perbarui tanggal setiap kali edit sesuai permintaan
	set["created_at"] = time.Now()
	update := bson.M{"$set": set}

	return produkCol().UpdateOne(ctx, bson.M{"_id": id}, update)
}

func DeleteProduk(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return produkCol().DeleteOne(ctx, bson.M{"_id": id})
}
