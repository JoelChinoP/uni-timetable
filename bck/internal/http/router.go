package http

import (
	"os"
	"time"

	/* auth "github.com/JoelChinoP/OAuth-with-Go/internal/auth"
	user "github.com/JoelChinoP/OAuth-with-Go/internal/user" */
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configura las rutas de la aplicación
func SetupRoutes(app *fiber.App) {

	// Hello World (raíz)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("¡Hello, World!")
	})

	// Configuración de CORS
	app.Get("/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"environment": os.Getenv("GO_ENV"),
			"version":     "1.0.0",
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	})

	/*** RUTAS LIBRES ***/
	/* 	auth.RegisterRoutes(app) // Ensure AuthRoutes is defined or imported
	 */
	/*** RUTAS PROTEGIDAS ***/
	/* 	api := app.Group("/api")
	   	api.Use(AuthenticationMiddleware()) // Middleware de autenticación para todas las rutas de la API
	*/
	// Rutas de la API
	/* 	user.RegisterRoutes(api) // Ensure UserRoutes is defined or imported */
}
