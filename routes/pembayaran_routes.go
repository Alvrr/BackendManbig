package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func PembayaranRoutes(app *fiber.App) {
	pembayaran := app.Group("/pembayaran")

	// GET semua pembayaran - admin, kasir
	pembayaran.Get("/", middleware.RoleGuard("admin", "kasir"), controllers.GetAllPembayaran)

	// GET by ID - admin, kasir
	pembayaran.Get("/:id", middleware.RoleGuard("admin", "kasir"), controllers.GetPembayaranByID)

	// POST - hanya admin, kasir
	pembayaran.Post("/", middleware.RoleGuard("admin", "kasir"), controllers.CreatePembayaran)

	// PUT selesaikan - admin, kasir
	pembayaran.Put("/selesaikan/:id", middleware.RoleGuard("admin", "kasir"), controllers.SelesaikanPembayaran)

	// Cetak surat jalan dihapus; dikelola oleh modul pengiriman
}
