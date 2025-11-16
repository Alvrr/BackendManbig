package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func PembayaranRoutes(app *fiber.App) {
	pembayaran := app.Group("/pembayaran")

	// GET semua pembayaran - admin, kasir, driver
	pembayaran.Get("/", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetAllPembayaran)

	// GET by ID - admin, kasir, driver
	pembayaran.Get("/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetPembayaranByID)

	// POST - hanya admin, kasir
	pembayaran.Post("/", middleware.RoleGuard("admin", "kasir"), controllers.CreatePembayaran)

	// PUT selesaikan - admin, kasir, driver
	pembayaran.Put("/selesaikan/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.SelesaikanPembayaran)

	// GET cetak surat jalan - admin, kasir, driver
	pembayaran.Get("/cetak/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.CetakSuratJalan)
}
