package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

// AuthenticatorJWT gera e valida access tokens HS256.
type AuthenticatorJWT struct {
	secret    []byte
	accessTTL time.Duration
}

// NewAuthenticatorJWT monta o autenticador com o secret HS256 e o TTL do
// access token. secret é lido de Config.JWTSecret; accessTTL de
// Config.JWTAccessTokenTTLMinutes.
func NewAuthenticatorJWT(secret string, accessTTL time.Duration) *AuthenticatorJWT {
	return &AuthenticatorJWT{secret: []byte(secret), accessTTL: accessTTL}
}

// GerarAccessToken assina um novo access token JWT (HS256) pro par
// usuário/papel, com expiração em now+accessTTL.
func (a *AuthenticatorJWT) GerarAccessToken(subject string, tipo domainauth.TipoUsuario, papel domainauth.PapelUsuario) (string, error) {
	now := time.Now()
	claims := domainauth.AppClaims{
		Subject: subject,
		Tipo:    tipo,
		Papel:   papel,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secret)
}

// GerarRefreshToken delega pra função de mesmo nome em refresh_hash.go —
// método existe só pra satisfazer domainauth.JWTProvider (mockável nos use cases).
func (a *AuthenticatorJWT) GerarRefreshToken() (string, error) {
	return GerarRefreshToken()
}

// ValidarAccessToken verifica assinatura HS256 e expiração do token bruto e,
// se válido, devolve os claims tipados. Rejeita qualquer método de assinatura
// diferente de HMAC.
func (a *AuthenticatorJWT) ValidarAccessToken(tokenBruto string) (*domainauth.AppClaims, error) {
	claims := &domainauth.AppClaims{}
	token, err := jwt.ParseWithClaims(tokenBruto, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metodo de assinatura inesperado")
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token invalido")
	}
	return claims, nil
}
