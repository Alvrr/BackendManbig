package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	ProdukRoutes(app)
	KategoriRoutes(app)
	StokRoutes(app)
	TransaksiRoutes(app)
	PelangganRoutes(app)
	PembayaranRoutes(app)
	PengirimanRoutes(app)
	LaporanRoutes(app)
	AuthRoutes(app)
	UserRoutes(app)
	RiwayatRoutes(app)
}
