package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword toma una contraseña y la hashea usando bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12) //14 es el costo recomendado
	return string(bytes), err
}

// CheckPasswordHash compara una contraseña con su hash usando bcrypt
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
