package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AppClaims são os claims do access token JWT.
// Subject e RegisteredClaims.Subject carregam o mesmo valor (ID do usuário);
// Tipo não tem claim registrado equivalente e viaja como campo custom.
// Jti identifica esse access token especificamente (não confundir com o
// refresh token bruto)
type AppClaims struct {
	Subject string              `json:"sub"`
	Tipo    TipoUsuario         `json:"tipo"`
	Papel   shared.PapelUsuario `json:"papel"`
	Jti     string              `json:"jti"`
	jwt.RegisteredClaims
}
