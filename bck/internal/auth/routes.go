package auth

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes configura las rutas de autenticación
func RegisterRoutes(router fiber.Router) {
	authGroup := router.Group("/auth")

	handler := NewHandler()

	// Rutas de autenticación
	authGroup.Post("/google/callback", handler.GoogleCallbackHandler)
	authGroup.Post("/login", handler.LoginHandler)
}
