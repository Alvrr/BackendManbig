package routes

import (
	"backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	user := app.Group("/users")
	user.Get("/drivers", controllers.GetAllDrivers)

	// CRUD karyawan (admin only)
	user.Get("/karyawan", controllers.GetAllKaryawan)
	user.Get("/karyawan/active", controllers.GetActiveKaryawan)
	user.Get("/karyawan/:id", controllers.GetKaryawanByID)
	user.Post("/karyawan", controllers.CreateKaryawan)
	user.Put("/karyawan/:id", controllers.UpdateKaryawan)
	user.Delete("/karyawan/:id", controllers.DeleteKaryawan)
	user.Patch("/karyawan/:id/status", controllers.UpdateKaryawanStatus)

	// Register karyawan (bisa dipakai di halaman karyawan, bukan login)
	user.Post("/register-karyawan", controllers.RegisterKaryawan)
}
