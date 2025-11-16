package routes

import (
	"backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func RiwayatRoutes(app *fiber.App) {
	riwayat := app.Group("/riwayat")
	riwayat.Get("/", controllers.GetRiwayatPembayaran)
}
