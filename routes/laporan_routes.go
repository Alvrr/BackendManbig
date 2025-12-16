package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LaporanRoutes(app *fiber.App) {
	app.Get("/laporan/export/excel", middleware.RoleGuard("admin"), controllers.ExportLaporanExcel)
	// Best sellers: bisa diakses admin dan gudang
	app.Get("/laporan/best-sellers", middleware.RoleGuard("admin", "gudang"), controllers.GetBestSellers)
}
