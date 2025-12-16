package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func StokRoutes(app *fiber.App) {
	g := app.Group("/stok")
	// View saldo & mutasi: semua role (admin, kasir, gudang, driver)
	g.Get("/saldo/:produk_id", middleware.RoleGuard("admin", "kasir", "gudang", "driver"), controllers.GetSaldoProduk)
	g.Get("/mutasi/:produk_id", middleware.RoleGuard("admin", "kasir", "gudang", "driver"), controllers.GetMutasiByProduk)
	// Create mutasi: admin + gudang
	g.Post("/", middleware.RoleGuard("admin", "gudang"), controllers.CreateMutasi)
}
