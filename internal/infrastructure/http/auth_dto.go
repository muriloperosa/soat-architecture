package http

// LoginRequest é o corpo HTTP de POST /v1/auth/login e /v1/auth/cliente/login.
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required"`
}

// RefreshRequest é o corpo HTTP de POST /v1/auth/refresh e /v1/auth/cliente/refresh,
// e do logout (mesmo formato, revoga em vez de rotacionar).
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TokenResponse é a resposta comum de login e refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
