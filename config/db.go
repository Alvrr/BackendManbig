package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global variabel untuk koleksi
var (
	DB                   *mongo.Database
	ProdukCollection     *mongo.Collection
	KategoriCollection   *mongo.Collection
	PelangganCollection  *mongo.Collection
	PembayaranCollection *mongo.Collection
	PengirimanCollection *mongo.Collection
	CounterCollection    *mongo.Collection
	UserCollection       *mongo.Collection
)

func ConnectDB() {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	// Defaults for local development if env not set
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	if dbName == "" {
		dbName = "mbg"
	}

	fmt.Println("✅ MONGO_URI:", mongoURI)
	fmt.Println("✅ DB_NAME:", dbName)

	// Setup client MongoDB
	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal connect ke MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ MongoDB tidak bisa diakses:", err)
	}

	fmt.Println("✅ Terhubung ke MongoDB")

	DB = client.Database(dbName)

	// Inisialisasi semua koleksi
	ProdukCollection = DB.Collection("produk")
	KategoriCollection = DB.Collection("kategori")
	PelangganCollection = DB.Collection("pelanggan")
	PembayaranCollection = DB.Collection("pembayaran")
	PengirimanCollection = DB.Collection("pengiriman")
	CounterCollection = DB.Collection("counters")
	UserCollection = DB.Collection("user")

	fmt.Println("✅ Semua koleksi berhasil diinisialisasi")
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
