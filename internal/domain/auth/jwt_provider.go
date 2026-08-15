package auth

// JWTProvider emite o access token JWT assinado pro par usuário/papel.
// Implementado em internal/infrastructure/auth (AutenticadorJWT).
type JWTProvider interface {
	GerarAccessToken(subject string, tipo TipoUsuario, papel PapelUsuario) (string, error)
}
