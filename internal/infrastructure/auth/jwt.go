package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AuthenticatorJWT gera e valida access tokens HS256, e gera o valor bruto
// do refresh token.
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
// usuário/papel, com expiração em now+accessTTL, e gera o jti que o
// identifica (usado pra revogação em par com o refresh token).
func (a *AuthenticatorJWT) GerarAccessToken(subject string, tipo domainauth.TipoUsuario, papel shared.PapelUsuario) (token string, jti string, err error) {
	jti, err = a.GerarRefreshToken()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	claims := domainauth.AppClaims{
		Subject: subject,
		Tipo:    tipo,
		Papel:   papel,
		Jti:     jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.accessTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = tok.SignedString(a.secret)
	if err != nil {
		return "", "", err
	}
	return token, jti, nil
}

// GerarRefreshToken gera um valor aleatório de alta entropia (32 bytes)
// codificado em base64, pra ser entregue ao cliente.
func (a *AuthenticatorJWT) GerarRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
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

// HashRefreshToken calcula o hash (SHA-256) do token bruto pra persistência
// e busca, nunca o valor bruto é armazenado.
func HashRefreshToken(bruto string) string {
	sum := sha256.Sum256([]byte(bruto))
	return hex.EncodeToString(sum[:])
}
