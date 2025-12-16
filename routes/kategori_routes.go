package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func KategoriRoutes(app *fiber.App) {
	g := app.Group("/kategori")

	// Semua role (dengan JWT) bisa melihat
	g.Get("/", controllers.GetAllKategori)
	g.Get("/:id", controllers.GetKategoriByID)

	// Hanya admin & gudang yang bisa create/update/delete
	g.Post("/", middleware.RoleGuard("admin", "gudang"), controllers.CreateKategori)
	g.Put("/:id", middleware.RoleGuard("admin", "gudang"), controllers.UpdateKategori)
	g.Delete("/:id", middleware.RoleGuard("admin", "gudang"), controllers.DeleteKategori)
}
