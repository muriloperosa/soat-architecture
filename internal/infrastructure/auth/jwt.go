package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

// AutenticadorJWT gera e valida access tokens HS256.
type AutenticadorJWT struct {
	secret    []byte
	accessTTL time.Duration
}

func NewAuthenticatorJWT(secret string, accessTTL time.Duration) *AutenticadorJWT {
	return &AutenticadorJWT{secret: []byte(secret), accessTTL: accessTTL}
}

func (a *AutenticadorJWT) GerarAccessToken(subject string, tipo domainauth.TipoUsuario, papel domainauth.PapelUsuario) (string, error) {
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
func (a *AutenticadorJWT) GerarRefreshToken() (string, error) {
	return GerarRefreshToken()
}

func (a *AutenticadorJWT) ValidarAccessToken(tokenBruto string) (*domainauth.AppClaims, error) {
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
