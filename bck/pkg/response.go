package pkg

import (
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse es la estructura para respuestas de éxito estandarizadas
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success genera una respuesta de éxito en formato JSON
func Success(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessWithStatus genera una respuesta de éxito con un código de estado específico
func SuccessWithStatus(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created genera una respuesta de éxito con un código de estado 201 (Creado)
func Created(c *fiber.Ctx, message string, data interface{}) error {
	return SuccessWithStatus(c, fiber.StatusCreated, message, data)
}

// Deleted genera una respuesta de éxito con un código de estado 204 (Sin contenido)
func Deleted(c *fiber.Ctx, message string, data interface{}) error {
	return SuccessWithStatus(c, fiber.StatusNoContent, message, data)
}