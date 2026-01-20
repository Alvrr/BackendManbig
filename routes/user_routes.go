package routes

import (
	"backend/controllers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	user := app.Group("/users")
	// List driver untuk kebutuhan mapping/pemilihan (admin & kasir)
	user.Get("/drivers", middleware.RoleGuard("admin", "kasir"), controllers.GetAllDrivers)

	// CRUD karyawan (admin only)
	user.Get("/karyawan", middleware.RoleGuard("admin"), controllers.GetAllKaryawan)
	user.Get("/karyawan/active", middleware.RoleGuard("admin"), controllers.GetActiveKaryawan)
	user.Get("/karyawan/:id", middleware.RoleGuard("admin"), controllers.GetKaryawanByID)
	user.Post("/karyawan", middleware.RoleGuard("admin"), controllers.CreateKaryawan)
	user.Put("/karyawan/:id", middleware.RoleGuard("admin"), controllers.UpdateKaryawan)
	user.Delete("/karyawan/:id", middleware.RoleGuard("admin"), controllers.DeleteKaryawan)
	user.Patch("/karyawan/:id/status", middleware.RoleGuard("admin"), controllers.UpdateKaryawanStatus)

	// Register karyawan (bisa dipakai di halaman karyawan, bukan login)
	user.Post("/register-karyawan", middleware.RoleGuard("admin"), controllers.RegisterKaryawan)
}
