package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	dto "github.com/JoelChinoP/OAuth-with-Go/internal/auth/dto"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// InitGoogleProvider initializes the Google OAuth provider with the necessary credentials.
var (
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"profile", "email"}, // Adjust scopes as needed
		Endpoint:     google.Endpoint,
	}

	// httpClient is a reusable HTTP client for making requests to Google APIs.
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

// GetUserInfo retrieves user information from Google using the access token.
func GetUserInfo(accessToken string) (*dto.GoogleUserInfo, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Revoke the token if it's expired
	defer revokeToken(accessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user info from Google")
	}

	var userInfo dto.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	//fmt.Println("User Info:", userInfo) // Debugging line to see the user info
	return &userInfo, nil
}

// revokeToken revokes the access token using Google's token revocation endpoint.
func revokeToken(token string) {
	resp, err := http.PostForm("https://oauth2.googleapis.com/revoke",
		url.Values{"token": {token}})

	if err != nil {
		fmt.Println("Error revoking token:", err)
		return
	}
	resp.Body.Close()
}
