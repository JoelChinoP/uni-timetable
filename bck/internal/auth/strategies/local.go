package auth

import (
	dto "github.com/JoelChinoP/OAuth-with-Go/internal/auth/dto"
	utils "github.com/JoelChinoP/OAuth-with-Go/internal/auth/utils"
)

// VerifyCredentials validates the user login credentials.
func VerifyCredentials(userLogin *dto.UserLogin, hashedPassword string) (bool, error) {

	isValid := utils.CheckPasswordHash(userLogin.Password, hashedPassword)
	if !isValid {
		return false, dto.ErrInvalidCredentials
	}
	return true, nil
}
