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

	// POST - kasir saja (admin read-only)
	pembayaran.Post("/", middleware.RoleGuard("kasir"), controllers.CreatePembayaran)

	// PUT selesaikan - kasir saja (admin read-only)
	pembayaran.Put("/selesaikan/:id", middleware.RoleGuard("kasir"), controllers.SelesaikanPembayaran)

	// Cetak surat jalan dihapus; dikelola oleh modul pengiriman
}
