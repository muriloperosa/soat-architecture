package shared_test

import (
	"strings"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewSenhaHash_Valida_GeraHashQueConfereComOriginal(t *testing.T) {
	s, err := shared.NewSenhaHash("senha123")
	require.NoError(t, err)
	require.True(t, s.Confere("senha123"))
	require.False(t, s.Confere("outra"))
	require.NotEqual(t, "senha123", s.String())
}

func TestNewSenhaHash_MenosDe8Caracteres_RetornaErro(t *testing.T) {
	_, err := shared.NewSenhaHash("1234567")
	require.ErrorIs(t, err, shared.ErrSenhaFraca)
}

func TestNewSenhaHash_MaisDe72Bytes_RetornaErroDoBcrypt(t *testing.T) {
	senhaLonga := strings.Repeat("a", 73)

	_, err := shared.NewSenhaHash(senhaLonga)

	require.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
}

func TestRestaurarSenhaHash_NaoRevalidaEConfereComHashOriginal(t *testing.T) {
	original, err := shared.NewSenhaHash("senha123")
	require.NoError(t, err)

	restaurada := shared.RestaurarSenhaHash(original.String())
	require.True(t, restaurada.Confere("senha123"))
}
