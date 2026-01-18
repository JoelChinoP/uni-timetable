package auth

import (
	"errors"

	dto "github.com/JoelChinoP/OAuth-with-Go/internal/auth/dto"
	"github.com/JoelChinoP/OAuth-with-Go/pkg"
	"github.com/gofiber/fiber/v2"
)

// Handler is the struct that contains the service for authentication
type Handler struct {
	service  *Service
	validate *pkg.Validator
}

// NewHandler creates a new instance of the handler
func NewHandler() *Handler {
	return &Handler{
		service:  NewAuthService(),
		validate: pkg.GetValidator(),
	}
}

// GoogleCallbackHandler handles the callback from Google after authentication.
func (h *Handler) GoogleCallbackHandler(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return errors.New("missing code parameter")
	}

	jwtToken, err := h.service.GoogleCallback(c, code)
	if err != nil {
		if errors.Is(err, dto.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid credentials",
			})
		}
		return errors.New("failed to exchange code for token")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User authenticated successfully",
		"token":   jwtToken,
	})
}

// LoginHandler handles user login and returns a JWT token if successful.
func (h *Handler) LoginHandler(c *fiber.Ctx) error {
	var req dto.UserLogin
	valid, error := h.validate.ValidateRequest(c, &req)
	if !valid {
		return h.validate.RespondWithValidationErrors(c, error)
	}

	jwtToken, err := h.service.LoginUser(c, &req)
	if err != nil {
		if errors.Is(err, dto.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid credentials",
			})
		}
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User authenticated successfully",
		"token":   jwtToken,
	})
}
