package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	dto "github.com/JoelChinoP/OAuth-with-Go/internal/auth/dto"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte(os.Getenv("JWT_SECRET"))

// SignGoogleJWT signs a JWT token with the provided Google user information and returns the signed token as a string.
func SignGoogleJWT(data *dto.Payload) (string, error) {
	signedToken, _ := buildToken(data)
	return signedToken, nil
}

// buildToken signs a JWT token with the provided user information and returns the signed token as a string.
func buildToken(user *dto.Payload) (string, error) {
	token := jwt.NewWithClaims( // Customize the claims as needed
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":    user.ID,
			"email": user.Email,
			//"role":  user.Role,
			"exp": time.Now().Add(time.Hour).Unix(), // dura1h
		},
	)

	signedToken, err := token.SignedString(secretKey) // Replace with your actual secret key
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// VerifyToken verifies the JWT token and returns the payload if valid.
func VerifyToken(tokenString string) (*dto.Payload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	// Extract claims correctamente
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	id, _ := claims["id"].(string)
	email, _ := claims["email"].(string)

	if id == "" || email == "" {
		return nil, errors.New("invalid token claims")
	}

	return &dto.Payload{
		ID:    id,
		Email: email,
	}, nil
}
