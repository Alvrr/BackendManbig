package middleware

import (
	"backend/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func JWTMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token tidak ditemukan atau format salah",
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := utils.ParseToken(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token tidak valid atau kadaluarsa",
		})
	}

	// Simpan claims ke context agar bisa dipakai di handler berikutnya
	c.Locals("userID", claims.ID)
	c.Locals("userRole", claims.Role)
	c.Locals("userNama", claims.Nama)

	return c.Next()
}
