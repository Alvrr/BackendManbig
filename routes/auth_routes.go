package routes

import (
	"backend/controllers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)

	// Endpoint untuk dropdown driver
	auth.Get("/drivers", controllers.GetAllDrivers)
}
