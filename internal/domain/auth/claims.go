package auth

import "github.com/golang-jwt/jwt/v5"

// AppClaims são os claims do access token JWT.
// Subject e RegisteredClaims.Subject carregam o mesmo valor (ID do usuário);
// Tipo não tem claim registrado equivalente e viaja como campo custom.
type AppClaims struct {
	Subject string       `json:"sub"`
	Tipo    TipoUsuario  `json:"tipo"`
	Papel   PapelUsuario `json:"papel"`
	jwt.RegisteredClaims
}
