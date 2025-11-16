package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func PelangganRoutes(app *fiber.App) {
	pelanggan := app.Group("/pelanggan")

	// GET bisa diakses admin, kasir, driver
	pelanggan.Get("/", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetAllPelanggan)
	pelanggan.Get("/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetPelangganByID)

	// POST/PUT/DELETE hanya admin, kasir
	pelanggan.Post("/", middleware.RoleGuard("admin", "kasir"), controllers.CreatePelanggan)
	pelanggan.Put("/:id", middleware.RoleGuard("admin", "kasir"), controllers.UpdatePelanggan)
	pelanggan.Delete("/:id", middleware.RoleGuard("admin", "kasir"), controllers.DeletePelanggan)
}
