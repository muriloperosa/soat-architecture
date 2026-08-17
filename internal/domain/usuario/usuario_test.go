package usuario_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/stretchr/testify/require"
)

func TestNewUsuario_Valido_NasceAtivoComRequerAlterarSenha(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	require.Equal(t, "Ana Souza", u.Nome())
	require.Equal(t, "ana@oficina.com", u.Email().String())
	require.Equal(t, shared.PapelMecanico, u.Papel())
	require.True(t, u.Ativo())
	require.True(t, u.RequerAlterarSenha())
	require.True(t, u.Senha().Confere("senha123"))
}

func TestNewUsuario_NomeVazio_RetornaErro(t *testing.T) {
	_, err := usuario.NewUsuario("", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.ErrorIs(t, err, usuario.ErrNomeObrigatorio)
}

func TestNewUsuario_PapelInvalido_RetornaErro(t *testing.T) {
	_, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelUsuario("gerente"))
	require.ErrorIs(t, err, usuario.ErrPapelInvalido)
}

func TestNewUsuario_EmailInvalido_RetornaErro(t *testing.T) {
	_, err := usuario.NewUsuario("Ana Souza", "nao-e-email", "senha123", shared.PapelMecanico)
	require.ErrorIs(t, err, shared.ErrEmailInvalido)
}

func TestNewUsuario_SenhaFraca_RetornaErro(t *testing.T) {
	_, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "curta", shared.PapelMecanico)
	require.ErrorIs(t, err, shared.ErrSenhaFraca)
}

func TestUsuario_AlterarSenha_EncerraRequerAlterarSenha(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.AlterarSenha("senhaNova123")
	require.NoError(t, err)
	require.False(t, u.RequerAlterarSenha())
	require.True(t, u.Senha().Confere("senhaNova123"))
	require.False(t, u.Senha().Confere("senha123"))
}

func TestUsuario_AlterarSenha_SenhaFraca_RetornaErroENaoAlteraSenhaAtual(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.AlterarSenha("curta")
	require.ErrorIs(t, err, shared.ErrSenhaFraca)
	require.True(t, u.RequerAlterarSenha())
	require.True(t, u.Senha().Confere("senha123"))
}

func TestUsuario_Atualizar_TrocaNomeEmailEPapel(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.Atualizar("Ana S. Costa", "ana.costa@oficina.com", shared.PapelAtendente)
	require.NoError(t, err)
	require.Equal(t, "Ana S. Costa", u.Nome())
	require.Equal(t, "ana.costa@oficina.com", u.Email().String())
	require.Equal(t, shared.PapelAtendente, u.Papel())
}

func TestUsuario_Atualizar_NomeVazio_RetornaErroENaoAltera(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.Atualizar("", "ana@oficina.com", shared.PapelAtendente)
	require.ErrorIs(t, err, usuario.ErrNomeObrigatorio)
	require.Equal(t, "Ana Souza", u.Nome())
	require.Equal(t, shared.PapelMecanico, u.Papel())
}

func TestUsuario_Atualizar_PapelInvalido_RetornaErroENaoAltera(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.Atualizar("Ana S. Costa", "ana@oficina.com", shared.PapelUsuario("gerente"))
	require.ErrorIs(t, err, usuario.ErrPapelInvalido)
	require.Equal(t, "Ana Souza", u.Nome())
	require.Equal(t, shared.PapelMecanico, u.Papel())
}

func TestUsuario_Atualizar_EmailInvalido_RetornaErroENaoAltera(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = u.Atualizar("Ana S. Costa", "nao-e-email", shared.PapelAtendente)
	require.ErrorIs(t, err, shared.ErrEmailInvalido)
	require.Equal(t, "Ana Souza", u.Nome())
	require.Equal(t, "ana@oficina.com", u.Email().String())
	require.Equal(t, shared.PapelMecanico, u.Papel())
}

func TestUsuario_RedefinirSenha_ForcaRequerAlterarSenha(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	require.NoError(t, u.AlterarSenha("senhaAntiga123"))
	require.False(t, u.RequerAlterarSenha())

	err = u.RedefinirSenha("senhaDoAdmin123")
	require.NoError(t, err)
	require.True(t, u.RequerAlterarSenha())
	require.True(t, u.Senha().Confere("senhaDoAdmin123"))
	require.False(t, u.Senha().Confere("senhaAntiga123"))
}

func TestUsuario_RedefinirSenha_SenhaFraca_RetornaErroENaoAlteraSenhaAtual(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	require.NoError(t, u.AlterarSenha("senhaAntiga123"))

	err = u.RedefinirSenha("curta")
	require.ErrorIs(t, err, shared.ErrSenhaFraca)
	require.False(t, u.RequerAlterarSenha())
	require.True(t, u.Senha().Confere("senhaAntiga123"))
}

func TestUsuario_AtivarInativar(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	u.Inativar()
	require.False(t, u.Ativo())

	u.Ativar()
	require.True(t, u.Ativo())
}

func TestUsuario_AtribuirID_PreencheID(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	require.Zero(t, u.ID())

	u.AtribuirID(7)
	require.Equal(t, uint64(7), u.ID())
}

func TestNewUsuario_DataCadastroEDataAtualizacao_NascemIguais(t *testing.T) {
	antes := time.Now()
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	depois := time.Now()

	require.False(t, u.DataCadastro().Before(antes))
	require.False(t, u.DataCadastro().After(depois))
	require.Equal(t, u.DataCadastro(), u.DataAtualizacao())
}

func TestUsuario_Atualizar_AtualizaDataAtualizacaoSemMudarDataCadastro(t *testing.T) {
	u, err := usuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	cadastroOriginal := u.DataCadastro()

	require.NoError(t, u.Atualizar("Ana S. Costa", "ana@oficina.com", shared.PapelAtendente))

	require.Equal(t, cadastroOriginal, u.DataCadastro())
	require.False(t, u.DataAtualizacao().Before(cadastroOriginal))
}

func TestRestaurarUsuario_NaoRevalidaEPreservaEstado(t *testing.T) {
	email, err := shared.NewEmail("ana@oficina.com")
	require.NoError(t, err)
	senha := shared.RestaurarSenhaHash("$2a$hash-existente")

	agora := time.Now()
	u := usuario.RestaurarUsuario(42, "Ana Souza", email, senha, shared.PapelMecanico, false, true, agora, agora)

	require.Equal(t, uint64(42), u.ID())
	require.Equal(t, "Ana Souza", u.Nome())
	require.False(t, u.RequerAlterarSenha())
	require.True(t, u.Ativo())
}
