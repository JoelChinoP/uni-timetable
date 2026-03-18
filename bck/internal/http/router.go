package http

import (
	"os"
	"strings"
	"time"

	"github.com/JoelChinoP/timetable_bck/internal/ui"
	"github.com/gofiber/fiber/v2"
)

const apiPrefix = "/api"

// SetupRoutes configures API routes and frontend delivery.
func SetupRoutes(app *fiber.App) {
	api := app.Group(apiPrefix)

	api.Get("/status", func(c *fiber.Ctx) error {
		environment := strings.TrimSpace(os.Getenv("GO_ENV"))
		if environment == "" {
			environment = "development"
		}

		return c.JSON(fiber.Map{
			"environment": environment,
			"version":     "1.0.0",
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	})

	/*** PUBLIC ROUTES ***/
	/* auth.RegisterRoutes(api) */
	/*** PROTECTED ROUTES ***/
	/* api.Use(AuthenticationMiddleware()) */
	/* user.RegisterRoutes(api) */

	ui.Mount(app, apiPrefix)
}
