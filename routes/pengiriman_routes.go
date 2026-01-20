package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func PengirimanRoutes(app *fiber.App) {
	g := app.Group("/pengiriman")
	// List & detail: admin, kasir, driver (driver sees own)
	g.Get("/", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetAllPengiriman)
	g.Get("/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetPengirimanByID)
	// Create: admin, kasir
	g.Post("/", middleware.RoleGuard("kasir"), controllers.CreatePengiriman)
	// Update: admin, kasir, driver (driver only own)
	g.Put("/:id", middleware.RoleGuard("kasir", "driver"), controllers.UpdatePengiriman)
	// Delete: admin, kasir
	g.Delete("/:id", middleware.RoleGuard("kasir"), controllers.DeletePengiriman)
}
