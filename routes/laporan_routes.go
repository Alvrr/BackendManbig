package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LaporanRoutes(app *fiber.App) {
	app.Get("/laporan/export/excel", middleware.RoleGuard("admin"), controllers.ExportLaporanExcel)
}
