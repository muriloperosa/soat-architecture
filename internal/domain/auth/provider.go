package auth

// JWTProvider emite e valida o access token JWT, e emite o valor bruto do
// refresh token. Implementado em internal/infrastructure/auth (AutenticadorJWT).
type JWTProvider interface {
	GerarAccessToken(subject string, tipo TipoUsuario, papel PapelUsuario) (string, error)
	ValidarAccessToken(tokenBruto string) (*AppClaims, error)
	GerarRefreshTokenBruto() (string, error)
}
