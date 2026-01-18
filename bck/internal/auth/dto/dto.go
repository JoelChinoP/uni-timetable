package auth

import "errors"

// ErrInvalidCredentials representa un error de credenciales inválidas
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserLogin contiene la información del usuario para el inicio de sesión
type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Payload contiene la información para el payload del JWT
type Payload struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// GoogleUserInfo contiene la información del usuario de Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}
