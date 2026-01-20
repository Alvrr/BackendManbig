package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func PelangganRoutes(app *fiber.App) {
	pelanggan := app.Group("/pelanggan")

	// GET bisa diakses admin, kasir, driver
	pelanggan.Get("/", middleware.RoleGuard("admin", "kasir"), controllers.GetAllPelanggan)
	pelanggan.Get("/:id", middleware.RoleGuard("admin", "kasir"), controllers.GetPelangganByID)

	// POST/PUT/DELETE hanya admin, kasir
	pelanggan.Post("/", middleware.RoleGuard("kasir"), controllers.CreatePelanggan)
	pelanggan.Put("/:id", middleware.RoleGuard("kasir"), controllers.UpdatePelanggan)
	pelanggan.Delete("/:id", middleware.RoleGuard("kasir"), controllers.DeletePelanggan)
}
