package shared

import "golang.org/x/crypto/bcrypt"

const senhaMinLength = 8

var ErrSenhaFraca = NewValidationError("senha deve ter ao menos 8 caracteres")

// SenhaHash é o VO de senha — sempre armazenado como hash bcrypt, a senha em
// texto plano nunca é retida após NewSenhaHash. Reusado por qualquer domínio
// com login por senha (usuario interno, cliente).
type SenhaHash struct {
	hash string
}

// NewSenhaHash valida a senha em texto plano (mínimo 8 caracteres) e retorna
// o VO já hasheado (bcrypt).
func NewSenhaHash(senhaPlana string) (SenhaHash, error) {
	if len(senhaPlana) < senhaMinLength {
		return SenhaHash{}, ErrSenhaFraca
	}
	h, err := bcrypt.GenerateFromPassword([]byte(senhaPlana), bcrypt.DefaultCost)
	if err != nil {
		return SenhaHash{}, NewInternalError("erro ao gerar hash de senha", err)
	}
	return SenhaHash{hash: string(h)}, nil
}

// RestaurarSenhaHash reidrata o VO a partir de um hash já persistido. Não
// revalida força de senha nem gera novo hash. Usado só por mappers de
// persistência.
func RestaurarSenhaHash(hash string) SenhaHash {
	return SenhaHash{hash: hash}
}

// Confere indica se senhaPlana corresponde ao hash armazenado.
func (s SenhaHash) Confere(senhaPlana string) bool {
	return bcrypt.CompareHashAndPassword([]byte(s.hash), []byte(senhaPlana)) == nil
}

// String retorna o hash bcrypt (pra persistência).
func (s SenhaHash) String() string {
	return s.hash
}
