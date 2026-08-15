package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GerarRefreshTokenBruto gera um valor aleatório de alta entropia (32 bytes)
// codificado em base64, pra ser entregue ao cliente.
func GerarRefreshTokenBruto() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefreshToken calcula o hash (SHA-256) do token bruto pra persistência
// e busca — nunca o valor bruto é armazenado.
func HashRefreshToken(bruto string) string {
	sum := sha256.Sum256([]byte(bruto))
	return hex.EncodeToString(sum[:])
}
