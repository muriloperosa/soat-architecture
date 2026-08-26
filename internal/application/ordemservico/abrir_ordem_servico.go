package ordemservico

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
)

type AbrirOrdemServicoUseCase struct {
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	clienteRepository      domaincliente.ClienteRepository
	veiculoRepository      domainveiculo.Repository
}

func NewAbrirOrdemServicoUseCase(
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
	clienteRepository domaincliente.ClienteRepository,
	veiculoRepository domainveiculo.Repository,
) *AbrirOrdemServicoUseCase {
	return &AbrirOrdemServicoUseCase{
		ordemServicoRepository: ordemServicoRepository,
		clienteRepository:      clienteRepository,
		veiculoRepository:      veiculoRepository,
	}
}

func (uc *AbrirOrdemServicoUseCase) Executar(
	ctx context.Context,
	input AbrirOrdemServicoInput,
) (OrdemServicoOutput, error) {
	cliente, err := uc.clienteRepository.BuscarPorID(ctx, input.ClienteID)
	if err != nil {
		return OrdemServicoOutput{}, err
	}
	if cliente == nil {
		return OrdemServicoOutput{}, domaincliente.ErrClienteNaoEncontrado
	}

	veiculo, err := uc.veiculoRepository.BuscarPorID(ctx, input.VeiculoID)
	if err != nil {
		return OrdemServicoOutput{}, err
	}
	if veiculo == nil {
		return OrdemServicoOutput{}, domainveiculo.ErrVeiculoNaoEncontrado
	}

	numero, err := gerarNumeroOrdemServico()
	if err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao gerar número da ordem de serviço", err)
	}

	os, err := domainordemservico.NewOrdemServico(
		numero,
		input.ClienteID,
		input.VeiculoID,
		int(input.QuilometragemEntrada),
		"",
		input.Observacoes,
		input.UsuarioID,
	)
	if err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.ordemServicoRepository.Salvar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao salvar ordem de serviço", err)
	}

	return toOutput(os), nil
}

func gerarNumeroOrdemServico() (string, error) {
	token := make([]byte, 6)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return "OS-" + time.Now().UTC().Format("20060102") + "-" + hex.EncodeToString(token), nil
}
