package auth

import (
	"log"

	dto "github.com/JoelChinoP/OAuth-with-Go/internal/auth/dto"
	strategies "github.com/JoelChinoP/OAuth-with-Go/internal/auth/strategies"
	"github.com/JoelChinoP/OAuth-with-Go/internal/db"
	sqlc "github.com/JoelChinoP/OAuth-with-Go/internal/db/sqlc"
	"github.com/gofiber/fiber/v2"
)

// Service encapsulates the operations for authentication services.
type Service struct {
	queries *sqlc.Queries
}

// NewAuthService creates a new instance of the authentication service.
func NewAuthService() *Service {
	sqlite := db.GetDB()
	return &Service{
		queries: sqlc.New(sqlite),
	}
}

// GoogleCallback handles the callback from Google after authentication.
func (s *Service) GoogleCallback(c *fiber.Ctx, code string) (string, error) {
	ctx := c.Context()

	// Intercambio de código por token
	token, err := strategies.GoogleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Printf("Error al intercambiar código por token: %v", err)
		return "", err
	}

	// Obtener datos del usuario de Google
	userInfo, err := strategies.GetUserInfo(token.AccessToken)
	if err != nil {
		return "", dto.ErrInvalidCredentials
	}

	// Buscar usuario por email y manejar error específicamente
	user, err := s.queries.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return "", dto.ErrInvalidCredentials
	}

	// Generar y retornar JWT
	return strategies.SignGoogleJWT(&dto.Payload{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}

// LoginUser handles user login and returns a JWT token if successful.
func (s *Service) LoginUser(c *fiber.Ctx, userLogin *dto.UserLogin) (string, error) {
	ctx := c.Context()

	// Obtener usuario por correo electrónico
	user, err := s.queries.GetUserByEmail(ctx, userLogin.Email)
	if err != nil {
		return "", dto.ErrInvalidCredentials
	}

	// Verificar credenciales
	isValid, err := strategies.VerifyCredentials(userLogin, user.HashedPassword)
	if err != nil || !isValid {
		return "", err
	}

	// Generar y retornar JWT
	return strategies.SignGoogleJWT(&dto.Payload{
		ID:    user.ID.String(),
		Email: user.Email,
	})
}
