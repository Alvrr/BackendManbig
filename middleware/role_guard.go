package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RoleGuard mengatur hak akses berdasarkan role
func RoleGuard(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Token tidak valid"})
		}
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak"})
	}
}
