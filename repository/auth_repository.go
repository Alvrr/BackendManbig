package repository

import (
	"context"
	"time"

	"backend/config"
	"backend/models"

	"go.mongodb.org/mongo-driver/bson"
)

// Cari user berdasarkan ID
func FindUserByID(id string) (*models.User, error) {
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("user")
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Buat user baru
func CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("user")

	// Pastikan field created_at diisi
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := collection.InsertOne(ctx, user)
	return err
}

// Cari user berdasarkan email
func FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("user")
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
