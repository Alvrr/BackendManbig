package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	ProdukRoutes(app)
	PelangganRoutes(app)
	PembayaranRoutes(app)
	LaporanRoutes(app)
	AuthRoutes(app)
	UserRoutes(app)
	RiwayatRoutes(app)
}
