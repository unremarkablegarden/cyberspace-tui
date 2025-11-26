package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	firebaseSignInURL    = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword"
	firebaseRefreshURL   = "https://securetoken.googleapis.com/v1/token"
	firebaseUserDataURL  = "https://identitytoolkit.googleapis.com/v1/accounts:lookup"
)

// AuthRequest is the request body for Firebase sign-in
type AuthRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

// AuthResponse is the response from Firebase sign-in
type AuthResponse struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalID      string `json:"localId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
}

// AuthError represents a Firebase auth error
type AuthError struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// RefreshRequest is the request body for token refresh
type RefreshRequest struct {
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse is the response from token refresh
type RefreshResponse struct {
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	UserID       string `json:"user_id"`
}

// SignIn authenticates with Firebase using email and password
func SignIn(email, password, apiKey string) (*AuthResponse, error) {
	reqBody := AuthRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?key=%s", firebaseSignInURL, apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var authErr AuthError
		if err := json.Unmarshal(body, &authErr); err != nil {
			return nil, fmt.Errorf("auth failed: %s", string(body))
		}
		return nil, fmt.Errorf("%s", friendlyError(authErr.Error.Message))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, err
	}

	return &authResp, nil
}

// RefreshToken exchanges a refresh token for a new ID token
func RefreshToken(refreshToken, apiKey string) (*RefreshResponse, error) {
	data := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)

	url := fmt.Sprintf("%s?key=%s", firebaseRefreshURL, apiKey)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var refreshResp RefreshResponse
	if err := json.Unmarshal(body, &refreshResp); err != nil {
		return nil, err
	}

	return &refreshResp, nil
}

// friendlyError converts Firebase error codes to user-friendly messages
func friendlyError(code string) string {
	switch code {
	case "EMAIL_NOT_FOUND":
		return "Email not found"
	case "INVALID_PASSWORD":
		return "Invalid password"
	case "USER_DISABLED":
		return "Account has been disabled"
	case "INVALID_EMAIL":
		return "Invalid email format"
	case "INVALID_LOGIN_CREDENTIALS":
		return "Invalid email or password"
	default:
		return code
	}
}
