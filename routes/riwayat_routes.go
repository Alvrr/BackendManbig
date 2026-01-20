package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func RiwayatRoutes(app *fiber.App) {
	riwayat := app.Group("/riwayat")
	// Riwayat pembayaran: admin (monitoring), kasir (milik sendiri), driver (milik sendiri)
	riwayat.Get("/", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetRiwayatPembayaran)
}
