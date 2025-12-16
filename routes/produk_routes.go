package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func ProdukRoutes(app *fiber.App) {
	produk := app.Group("/produk")

	// GET bisa diakses semua role (admin, kasir, gudang, driver)
	produk.Get("/", middleware.RoleGuard("admin", "kasir", "gudang", "driver"), controllers.GetAllProduk)
	produk.Get("/:id", middleware.RoleGuard("admin", "kasir", "gudang", "driver"), controllers.GetProdukByID)

	// POST/PUT/DELETE hanya admin, gudang
	produk.Post("/", middleware.RoleGuard("admin", "gudang"), controllers.CreateProduk)
	produk.Put("/:id", middleware.RoleGuard("admin", "gudang"), controllers.UpdateProduk)
	produk.Delete("/:id", middleware.RoleGuard("admin", "gudang"), controllers.DeleteProduk)
}
