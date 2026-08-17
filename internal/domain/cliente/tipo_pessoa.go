package cliente

type TipoPessoa string

const (
	TipoPessoaFisica   TipoPessoa = "PF"
	TipoPessoaJuridica TipoPessoa = "PJ"
)

func (t TipoPessoa) IsValid() bool {
	return t == TipoPessoaFisica || t == TipoPessoaJuridica
}