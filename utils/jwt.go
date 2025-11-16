package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Struktur custom claims untuk JWT
type JWTClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Nama string `json:"nama"` // Nama user ditambahkan ke dalam token
	jwt.RegisteredClaims
}

// Fungsi untuk menghasilkan JWT token
func GenerateToken(id, role, nama string) (string, error) {
	claims := JWTClaims{
		ID:   id,
		Role: role,
		Nama: nama,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)), // expired dalam 8 jam
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Ambil secret dari env saat dibutuhkan
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET tidak ditemukan di environment")
	}

	return token.SignedString([]byte(jwtSecret))
}

// Fungsi untuk memverifikasi dan parsing token
func ParseToken(tokenStr string) (*JWTClaims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET tidak ditemukan di environment")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}
