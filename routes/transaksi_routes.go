package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func TransaksiRoutes(app *fiber.App) {
	r := app.Group("/transaksi")
	// Read-only monitoring: admin bisa view; kasir hanya miliknya
	r.Get("/", middleware.RoleGuard("admin", "kasir"), controllers.ListTransaksi)
	r.Get("/:id", middleware.RoleGuard("admin", "kasir"), controllers.GetTransaksiByID)
	// Write: kasir saja
	r.Post("/", middleware.RoleGuard("kasir"), controllers.CreateTransaksi)
	r.Put("/:id", middleware.RoleGuard("kasir"), controllers.UpdateTransaksi)
	r.Delete("/:id", middleware.RoleGuard("kasir"), controllers.DeleteTransaksi)
}
