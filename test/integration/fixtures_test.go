//go:build integration

package integration_test

import (
	"context"
	"testing"

	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// resetDB limpa as tabelas usadas pelos testes de auth/usuario entre um teste e outro
func resetDB(t *testing.T) {
	t.Helper()
	require := func(err error) {
		if err != nil {
			t.Fatalf("erro ao limpar banco de teste: %v", err)
		}
	}
	require(testDB.Exec("DELETE FROM refresh_tokens").Error)
	require(testDB.Exec("DELETE FROM clientes").Error)
	require(testDB.Exec("DELETE FROM usuarios").Error)
}

// seedUsuario cria um usuário interno direto via use case (mesmo caminho do
// cmd/create-user), sem depender de rota HTTP.
func seedUsuario(t *testing.T, nome, email, senha string, papel shared.PapelUsuario) appusuario.UsuarioOutput {
	t.Helper()
	out, err := testContainer.CriarUsuarioUC.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome:         nome,
		Email:        email,
		SenhaInicial: senha,
		Papel:        papel,
	})
	if err != nil {
		t.Fatalf("erro ao semear usuário de teste: %v", err)
	}
	return out
}
