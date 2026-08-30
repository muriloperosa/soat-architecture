package ordemservico

import (
	"strconv"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func formatUint(value uint64) string {
	return strconv.FormatUint(value, 10)
}

func validarAcessoConsultaOrdemServico(
	ordemServico *domain.OrdemServico,
	solicitanteID uint64,
	tipoSolicitante domainauth.TipoUsuario,
) error {
	switch tipoSolicitante {
	case domainauth.TipoInterno:
		return nil
	case domainauth.TipoCliente:
		if solicitanteID == 0 {
			return shared.NewUnauthorizedError("requisição não autorizada")
		}

		if ordemServico.ClienteID() != solicitanteID {
			return shared.NewForbiddenError("acesso à ordem de serviço não permitido")
		}

		return nil
	default:
		return shared.NewUnauthorizedError("requisição não autorizada")
	}
}
