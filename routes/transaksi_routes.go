package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func TransaksiRoutes(app *fiber.App) {
	r := app.Group("/transaksi")
	// View: admin semua; kasir hanya miliknya
	r.Get("/", middleware.RoleGuard("admin", "kasir"), controllers.ListTransaksi)
	r.Get("/:id", middleware.RoleGuard("admin", "kasir", "driver"), controllers.GetTransaksiByID)
	// Write: admin+kasir (ownership checked in controller)
	r.Post("/", middleware.RoleGuard("admin", "kasir"), controllers.CreateTransaksi)
	r.Put("/:id", middleware.RoleGuard("admin", "kasir"), controllers.UpdateTransaksi)
	r.Delete("/:id", middleware.RoleGuard("admin", "kasir"), controllers.DeleteTransaksi)
}
