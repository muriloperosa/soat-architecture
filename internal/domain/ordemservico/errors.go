package ordemservico

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrClienteObrigatorio              = shared.NewValidationError("cliente é obrigatório")
	ErrVeiculoObrigatorio              = shared.NewValidationError("veículo é obrigatório")
	ErrNumeroObrigatorio               = shared.NewValidationError("número da ordem de serviço é obrigatório")
	ErrNumeroInvalido                  = shared.NewValidationError("número da ordem de serviço é inválido")
	ErrQuilometragemEntradaInvalida    = shared.NewValidationError("quilometragem de entrada é inválida")
	ErrStatusInvalido                  = shared.NewValidationError("status da ordem de serviço é inválido")
	ErrCriadoPorObrigatorio            = shared.NewValidationError("usuário responsável pela criação é obrigatório")
	ErrResponsavelHistoricoObrigatorio = shared.NewValidationError("usuário responsável pela alteração de status é obrigatório")
	ErrTransicaoStatusInvalida         = shared.NewValidationError("transição de status da ordem de serviço não permitida")
	ErrDiagnosticoObrigatorio          = shared.NewValidationError("diagnóstico é obrigatório")
	ErrDiagnosticoStatusInvalido       = shared.NewValidationError("diagnóstico só pode ser informado quando a ordem de serviço estiver em diagnóstico")
	ErrOrdemServicoNaoEncontrada       = shared.NewNotFoundError("ordem de serviço não encontrada")
)
