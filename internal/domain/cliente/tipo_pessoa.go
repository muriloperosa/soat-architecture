package cliente

type TipoPessoa string

const (
	TipoPessoaFisica   TipoPessoa = "PF"
	TipoPessoaJuridica TipoPessoa = "PJ"
)

func (t TipoPessoa) IsValid() bool {
	switch t {
	case TipoPessoaFisica, TipoPessoaJuridica:
		return true
	default:
		return false
	}
}
