package servico_test

import (
	"context"
	"errors"
	"testing"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func novoServico(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 1)
	require.NoError(t, err)
	return s
}

func TestCriarServicoUseCase_Executar_DadosValidos_CriaAtivo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { s.AtribuirID(1) }).
		Return(nil)

	out, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: 150.50, TempoEstimadoMinutos: 60, CriadoPor: 1,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.True(t, out.Ativo)
	require.Equal(t, 150.50, out.PrecoBase)
	require.Equal(t, 60, out.TempoEstimadoMinutos)
	require.Equal(t, uint64(1), out.CriadoPor)
}

func TestCriarServicoUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	_, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "", Descricao: "descrição", PrecoBase: 100, TempoEstimadoMinutos: 30, CriadoPor: 1,
	})

	require.ErrorIs(t, err, domainservico.ErrNomeObrigatorio)
}

func TestCriarServicoUseCase_Executar_ErroDoBancoAoSalvar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "Troca de óleo", Descricao: "descrição", PrecoBase: 100, TempoEstimadoMinutos: 30, CriadoPor: 1,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarServicoUseCase_Executar_ServicoExiste_AtualizaCampos(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	out, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "alinhamento e balanceamento", PrecoBase: 200.75, TempoEstimadoMinutos: 90,
	})

	require.NoError(t, err)
	require.Equal(t, "Alinhamento", out.Nome)
	require.Equal(t, 200.75, out.PrecoBase)
	require.Equal(t, 90, out.TempoEstimadoMinutos)
	require.Equal(t, uint64(1), out.CriadoPor)
	require.True(t, out.Ativo)
}

func TestAtualizarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 99, Nome: "X", Descricao: "Y", PrecoBase: 10, TempoEstimadoMinutos: 15,
	})

	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestAtualizarServicoUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarServicoUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	require.ErrorIs(t, err, domainservico.ErrNomeObrigatorio)
}

func TestAtualizarServicoUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestListarServicosUseCase_Executar_RetornaCatalogo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	s1 := novoServico(t)
	s1.AtribuirID(1)
	s2, err := domainservico.NewServico("Alinhamento", "alinhamento e balanceamento", 200, 90, 1)
	require.NoError(t, err)
	s2.AtribuirID(2)

	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{s1, s2}, nil)

	out, err := uc.Executar(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "Troca de óleo", out[0].Nome)
	require.Equal(t, "Alinhamento", out[1].Nome)
}

func TestListarServicosUseCase_Executar_ListaVazia_RetornaSliceVazio(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{}, nil)

	out, err := uc.Executar(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
	require.NotNil(t, out)
}

func TestListarServicosUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	repo.EXPECT().Listar(mock.Anything).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background())

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestBuscarServicoUseCase_Executar_ServicoExiste_RetornaDados(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	out, err := uc.Executar(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.Equal(t, "Troca de óleo", out.Nome)
}

func TestBuscarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	_, err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestBuscarServicoUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtivarServicoUseCase_Executar_AtivaServicoInativo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { require.True(t, s.Ativo()) }).
		Return(nil)

	require.NoError(t, uc.Executar(context.Background(), 1))
}

func TestAtivarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestAtivarServicoUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtivarServicoUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestInativarServicoUseCase_Executar_InativaServicoAtivo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewInativarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { require.False(t, s.Ativo()) }).
		Return(nil)

	require.NoError(t, uc.Executar(context.Background(), 1))
}

func TestInativarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewInativarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestInativarServicoUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewInativarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestInativarServicoUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewInativarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

