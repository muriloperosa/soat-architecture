package auth

// JWTProvider emite o access token JWT assinado e o valor bruto do refresh
// token. Implementado em internal/infrastructure/auth (AutenticadorJWT).
type JWTProvider interface {
	GerarAccessToken(subject string, tipo TipoUsuario, papel PapelUsuario) (string, error)
	GerarRefreshTokenBruto() (string, error)
}
