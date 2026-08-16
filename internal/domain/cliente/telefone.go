package cliente

import (
	"fmt"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

type Telefone struct {
	valor string
}

func NewTelefone(valor string) (Telefone, error) {
	valor = texts.OnlyNumbers(valor)

	if valor == "" {
		return Telefone{}, ErrTelefoneObrigatorio
	}

	if !validarTelefone(valor) {
		return Telefone{}, ErrTelefoneInvalido
	}

	return Telefone{valor: valor}, nil
}

func (t Telefone) String() string {
	return t.valor
}

func (t Telefone) Formatado() string {
	switch len(t.valor) {
	case 10:
		return fmt.Sprintf("(%s) %s-%s", t.valor[0:2], t.valor[2:6], t.valor[6:10])
	case 11:
		return fmt.Sprintf("(%s) %s-%s", t.valor[0:2], t.valor[2:7], t.valor[7:11])
	default:
		return t.valor
	}
}

func (t Telefone) IsZero() bool {
	return t.valor == ""
}

func validarTelefone(valor string) bool {
	if len(valor) != 10 && len(valor) != 11 {
		return false
	}

	// DDD deve possuir dois dígitos e não pode iniciar com zero.
	if valor[0] == '0' {
		return false
	}

	if len(valor) == 11 {
		return validarCelular(valor)
	}

	return validarTelefoneFixo(valor)
}

func validarCelular(valor string) bool {
	// Após o DDD, celulares brasileiros possuem 9 dígitos
	// e começam com 9.
	return valor[2] == '9'
}

func validarTelefoneFixo(valor string) bool {
	// Telefones fixos possuem 8 dígitos após o DDD.
	// Para o domínio do projeto restringimos o primeiro dígito
	// do número às faixas usuais de telefonia fixa.
	primeiroDigito := valor[2]

	return primeiroDigito >= '2' && primeiroDigito <= '5'
}
