package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LaporanRoutes(app *fiber.App) {

	laporanController := controllers.NewLaporanController()

	// Best sellers for dashboard (admin, gudang, kasir)
	app.Get(
		"/laporan/best-sellers",
		middleware.RoleGuard("admin", "gudang", "kasir"),
		laporanController.BestSellers,
	)

	app.Get(
		"/laporan/export/excel",
		middleware.JWTMiddleware,
		middleware.RoleGuard("admin"),
		laporanController.ExportExcel,
	)
}
