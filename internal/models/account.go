package models

// Basic user account data (shared)
type User struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// Response for /signup and /login
type AuthWithTokensResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Response for /auth
type AuthCheckResponse struct {
	User User `json:"user"`
}

// Response for /token/refresh
type TokenRefreshResponse struct {
	AccessToken string `json:"access_token"`
}
