package teacher

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/JoelChinoP/timetable_bck/pkg"
)

// TeacherHandler maneja las solicitudes HTTP relacionadas con los profesores
type TeacherHandler struct {
	service TeacherService
}

// GetTeacherHandler crea una nueva instancia del manejador de profesores
func GetTeacherHandler(service TeacherService) *TeacherHandler {
	return &TeacherHandler{
		service: service,
	}
}

// CreateUserHandler maneja la creación de un nuevo profesor
func (h *TeacherHandler) CreateUserHandler(c *fiber.Ctx) error {
	var req CreateTeacherDTO

	if errs := pkg.BindAndValidate(c, &req); len(errs) > 0 {
		return pkg.RespondValidation(c, errs)
	}

	teacher, err := h.service.CreateTeacher(c.Context(), &req)

	if err != nil {
		return err
	}
	return pkg.Created(c, fmt.Sprintf("Teacher created with ID %d", teacher.ID), teacher)
}

// ListUsersHandler maneja la obtención de la lista de profesores
func (h *TeacherHandler) ListUsersHandler(c *fiber.Ctx) error {
	teachers, err := h.service.ListTeachers(c.Context())
	if err != nil {
		return err
	}
	return pkg.Success(c, "Teachers retrieved successfully", teachers)
}

// GetUserByIDHandler maneja la obtención de un profesor por su ID
func (h *TeacherHandler) GetUserByIDHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	teacher, err := h.service.GetTeacherByID(c.Context(), int32(id))
	if err != nil {
		return err
	}
	return pkg.Success(c, "Teacher retrieved successfully", teacher)
}

// UpdateUserHandler maneja la actualización de un profesor existente
func (h *TeacherHandler) UpdateUserHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	var req UpdateTeacherDTO
	if errs := pkg.BindAndValidate(c, &req); len(errs) > 0 {
		return pkg.RespondValidation(c, errs)
	}

	teacher, err := h.service.UpdateTeacher(c.Context(), int32(id), &req)
	if err != nil {
		return err
	}
	return pkg.Success(c, "Teacher updated successfully", teacher)
}

// DeleteUserHandler maneja la eliminación de un profesor por su ID
func (h *TeacherHandler) DeleteUserHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	if err := h.service.DeleteTeacher(c.Context(), int32(id)); err != nil {
		return err
	}
	return pkg.Deleted(c, "Teacher deleted successfully", nil)
}
