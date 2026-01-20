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

func userCol() *mongo.Collection {
	return config.UserCollection
}

// EnsureUserIndexes ensures unique constraints for user identity fields.
// NOTE: If existing data contains duplicates, MongoDB will reject index creation.
func EnsureUserIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ci := &options.Collation{Locale: "en", Strength: 2} // case-insensitive
	_, err := userCol().Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetName("uniq_email_ci").SetUnique(true).SetCollation(ci),
		},
		{
			Keys:    bson.D{{Key: "nama", Value: 1}},
			Options: options.Index().SetName("uniq_nama_ci").SetUnique(true).SetCollation(ci),
		},
	})
	return err
}

func ExistsUserByEmail(email string, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"email": email}
	if excludeID != "" {
		filter["_id"] = bson.M{"$ne": excludeID}
	}

	ci := &options.Collation{Locale: "en", Strength: 2}
	err := userCol().FindOne(ctx, filter, options.FindOne().SetCollation(ci)).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func ExistsUserByNama(nama string, excludeID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"nama": nama}
	if excludeID != "" {
		filter["_id"] = bson.M{"$ne": excludeID}
	}

	ci := &options.Collation{Locale: "en", Strength: 2}
	err := userCol().FindOne(ctx, filter, options.FindOne().SetCollation(ci)).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// List all drivers
func GetAllDrivers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"role": "driver"}
	cursor, err := userCol().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []models.User
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		drivers = append(drivers, u)
	}
	return drivers, nil
}

// CRUD Karyawan (User, kecuali role admin)
func GetAllKaryawan() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"role": bson.M{"$ne": "admin"}}
	cursor, err := userCol().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func GetKaryawanByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var user models.User
	err := userCol().FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func CreateKaryawan(user *models.User) (*mongo.InsertOneResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return userCol().InsertOne(ctx, user)
}

func UpdateKaryawan(id string, user models.User) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"nama":   user.Nama,
			"email":  user.Email,
			"role":   user.Role,
			"no_hp":  user.NoHP,
			"alamat": user.Alamat,
			"status": user.Status,
		},
	}

	// Jika password tidak kosong, update juga password
	if user.Password != "" {
		update["$set"].(bson.M)["password"] = user.Password
	}

	return userCol().UpdateOne(ctx, bson.M{"_id": id}, update)
}

func DeleteKaryawan(id string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return userCol().DeleteOne(ctx, bson.M{"_id": id})
}

// Update status aktif/nonaktif karyawan
func UpdateKaryawanStatus(id string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{"$set": bson.M{"status": status}}
	result, err := userCol().UpdateOne(ctx, bson.M{"_id": id}, update)

	if err != nil {
		return err
	}

	// Periksa apakah ada dokumen yang diupdate
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

// Get all active karyawan (selain admin)
func GetActiveKaryawan() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.M{"status": "aktif", "role": bson.M{"$ne": "admin"}}
	cursor, err := userCol().Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var users []models.User
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
