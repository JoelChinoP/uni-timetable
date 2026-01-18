package teacher

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes registra las rutas de usuario en el router
func RegisterRoutes(router fiber.Router) {
	userGroup := router.Group("/users")

	// Dependencias
	service := GetTeacherService()
	handler := GetTeacherHandler(service)

	// Rutas de Usuarios
	userGroup.Post("", handler.CreateUserHandler)
	userGroup.Get("", handler.ListUsersHandler)
	userGroup.Get("/:id", handler.GetUserByIDHandler)
	userGroup.Put("/:id", handler.UpdateUserHandler)
	userGroup.Delete("/:id", handler.DeleteUserHandler)
}
