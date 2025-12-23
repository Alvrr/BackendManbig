package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func LaporanRoutes(app *fiber.App) {

	laporanController := controllers.NewLaporanController()

	app.Get(
		"/laporan/export/excel",
		middleware.JWTMiddleware,
		middleware.RoleGuard("admin"),
		laporanController.ExportExcel,
	)
}
