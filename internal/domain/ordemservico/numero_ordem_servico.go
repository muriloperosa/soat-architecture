package ordemservico

import "github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"

const tamanhoMaximoNumeroOrdemServico = 50

// NumeroOrdemServico é o identificador de negócio legível da OS.
type NumeroOrdemServico struct {
	valor string
}

func NewNumeroOrdemServico(valor string) (NumeroOrdemServico, error) {
	valor = texts.NormalizeSpaces(valor)
	if valor == "" {
		return NumeroOrdemServico{}, ErrNumeroObrigatorio
	}
	if len(valor) > tamanhoMaximoNumeroOrdemServico {
		return NumeroOrdemServico{}, ErrNumeroInvalido
	}

	return NumeroOrdemServico{valor: valor}, nil
}

func (n NumeroOrdemServico) String() string { return n.valor }
