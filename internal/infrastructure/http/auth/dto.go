package auth

// LoginRequest é o corpo HTTP de POST /v1/auth/login e /v1/auth/cliente/login.
type LoginRequest struct {
	Email string `json:"email" binding:"required,email" example:"usuario@oficina.com"`
	Senha string `json:"senha" binding:"required" example:"senha123"`
}

// RefreshRequest é o corpo HTTP de POST /v1/auth/refresh e /v1/auth/cliente/refresh,
// e do logout (mesmo formato, revoga em vez de rotacionar).
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"9f8c1e...b3a0"`
}

// TokenResponse é a resposta comum de login e refresh.
type TokenResponse struct {
	AccessToken           string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	RefreshToken          string `json:"refresh_token" example:"9f8c1e...b3a0"`
	ExpiresIn             int64  `json:"access_token_expires_in" example:"900"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in" example:"604800"`
}
