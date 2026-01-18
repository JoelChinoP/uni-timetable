package teacher

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	ErrTeacherNotFound = fiber.NewError(fiber.StatusNotFound, "Teacher not found")
	ErrTeacherExists   = fiber.NewError(fiber.StatusConflict, "Teacher with the same name and last name already exists")
)

type TeacherDTO struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

type CreateTeacherDTO struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	LastName string `json:"last_name" validate:"required,min=3,max=100"`
}

type UpdateTeacherDTO struct {
	Name     string `json:"name" validate:"min=3,max=100"`
	LastName string `json:"last_name" validate:"min=3,max=100"`
}
