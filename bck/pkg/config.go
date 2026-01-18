package pkg

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// ErrorResponse es la estructura para respuestas de error estandarizadas
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SetupCORS configura CORS para la aplicación Fiber
func SetupCORS(app *fiber.App, origins string) {
	if strings.TrimSpace(origins) == "" {
		origins = "*"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
}

// SetupGlobalErrorHandler configura el manejador de errores global para la aplicación Fiber
func SetupGlobalErrorHandler(c *fiber.Ctx, err error) error {
	// Si el error es de tipo *fiber.Error, se maneja de manera especial
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Success: false,
		Message: err.Error(),
	})
}
