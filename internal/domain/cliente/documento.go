package cliente

import "github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"

type Documento struct {
	valor string
	tipo  TipoPessoa
}

func NewDocumento(valor string, tipo TipoPessoa) (Documento, error) {
	valor = texts.OnlyNumbers(valor)

	if valor == "" {
		return Documento{}, ErrDocumentoObrigatorio
	}

	if !tipo.IsValid() {
		return Documento{}, ErrTipoPessoaInvalido
	}

	switch tipo {
	case TipoPessoaFisica:
		if !validarCPF(valor) {
			return Documento{}, ErrCPFInvalido
		}

	case TipoPessoaJuridica:
		if !validarCNPJ(valor) {
			return Documento{}, ErrCNPJInvalido
		}
	}

	return Documento{valor: valor, tipo: tipo}, nil
}

func (d Documento) String() string {
	return d.valor
}

func (d Documento) Tipo() TipoPessoa {
	return d.tipo
}

func (d Documento) Formatado() string {
	switch d.tipo {
	case TipoPessoaFisica:
		return formatarCPF(d.valor)

	case TipoPessoaJuridica:
		return formatarCNPJ(d.valor)

	default:
		return d.valor
	}
}

func (d Documento) IsZero() bool {
	return d.valor == ""
}

func validarCPF(valor string) bool {
	if len(valor) != 11 {
		return false
	}

	if todosDigitosIguais(valor) {
		return false
	}

	soma := 0
	for i := range 9 {
		soma += int(valor[i]-'0') * (10 - i)
	}

	resto := soma % 11
	digito1 := 11 - resto

	if digito1 >= 10 {
		digito1 = 0
	}

	if digito1 != int(valor[9]-'0') {
		return false
	}

	soma = 0
	for i := range 10 {
		soma += int(valor[i]-'0') * (11 - i)
	}

	resto = soma % 11
	digito2 := 11 - resto

	if digito2 >= 10 {
		digito2 = 0
	}

	return digito2 == int(valor[10]-'0')
}

func validarCNPJ(valor string) bool {
	if len(valor) != 14 {
		return false
	}

	if todosDigitosIguais(valor) {
		return false
	}

	pesosPrimeiro := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	soma := 0

	for i, peso := range pesosPrimeiro {
		soma += int(valor[i]-'0') * peso
	}

	resto := soma % 11

	digito1 := 0
	if resto >= 2 {
		digito1 = 11 - resto
	}

	if digito1 != int(valor[12]-'0') {
		return false
	}

	pesosSegundo := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	soma = 0

	for i, peso := range pesosSegundo {
		soma += int(valor[i]-'0') * peso
	}

	resto = soma % 11

	digito2 := 0
	if resto >= 2 {
		digito2 = 11 - resto
	}

	return digito2 == int(valor[13]-'0')
}

func todosDigitosIguais(valor string) bool {
	for i := 1; i < len(valor); i++ {
		if valor[i] != valor[0] {
			return false
		}
	}

	return true
}

func formatarCPF(valor string) string {
	if len(valor) != 11 {
		return valor
	}

	return valor[0:3] + "." + valor[3:6] + "." + valor[6:9] + "-" + valor[9:11]
}

func formatarCNPJ(valor string) string {
	if len(valor) != 14 {
		return valor
	}

	return valor[0:2] + "." + valor[2:5] + "." + valor[5:8] + "/" + valor[8:12] + "-" + valor[12:14]
}
