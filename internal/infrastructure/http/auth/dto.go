package auth

// LoginRequest é o corpo HTTP de POST /v1/auth/login e /v1/auth/cliente/login.
type LoginRequest struct {
	// Email cadastrado do usuário
	Email string `json:"email" binding:"required,email" example:"usuario@oficina.com"`
	// Senha em texto plano, validada contra o hash bcrypt armazenado
	Senha string `json:"senha" binding:"required" example:"senha123"`
}

// RefreshRequest é o corpo HTTP de POST /v1/auth/refresh e /v1/auth/cliente/refresh,
// e do logout (mesmo formato, revoga em vez de rotacionar).
type RefreshRequest struct {
	// RefreshToken bruto emitido em um login ou refresh anterior
	RefreshToken string `json:"refresh_token" binding:"required" example:"9f8c1e...b3a0"`
}

// TokenResponse é a resposta comum de login e refresh.
type TokenResponse struct {
	// AccessToken JWT (HS256) de curta duração, usado no header Authorization
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`
	// RefreshToken bruto de longa duração, revogado e rotacionado a cada uso
	RefreshToken string `json:"refresh_token" example:"9f8c1e...b3a0"`
}
