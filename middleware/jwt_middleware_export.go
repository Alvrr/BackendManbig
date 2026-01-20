package middleware

import (
	"backend/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddlewareForExport reads JWT from Authorization header.
// If missing, it also accepts a token from query string (default: ?token=...).
// This is intentionally scoped for download endpoints invoked via window.open.
func JWTMiddlewareForExport(c *fiber.Ctx) error {
	tokenStr := ""

	authHeader := c.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
	}

	if tokenStr == "" {
		tokenStr = c.Query("token", "")
	}

	if tokenStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token tidak ditemukan atau format salah",
		})
	}

	claims, err := utils.ParseToken(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token tidak valid atau kadaluarsa",
		})
	}

	c.Locals("userID", claims.ID)
	c.Locals("userRole", claims.Role)
	c.Locals("userNama", claims.Nama)

	return c.Next()
}
