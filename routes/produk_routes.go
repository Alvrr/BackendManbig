package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func ProdukRoutes(app *fiber.App) {
	produk := app.Group("/produk")

	// GET bisa diakses admin, kasir, driver
	produk.Get("/", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetAllProduk)
	produk.Get("/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetProdukByID)

	// POST/PUT/DELETE hanya admin, kasir
	produk.Post("/", middleware.RoleGuard("admin", "kasir"), controllers.CreateProduk)
	produk.Put("/:id", middleware.RoleGuard("admin", "kasir"), controllers.UpdateProduk)
	produk.Delete("/:id", middleware.RoleGuard("admin", "kasir"), controllers.DeleteProduk)
}
