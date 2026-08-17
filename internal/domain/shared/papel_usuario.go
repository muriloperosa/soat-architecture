package shared

// PapelUsuario distingue o papel do usuário dentro do seu tipo de login.
// Usuários internos têm um dos papéis admin/mecanico/atendente;
// usuários cliente têm sempre PapelCliente.
type PapelUsuario string

const (
	PapelAdmin     PapelUsuario = "ADMINISTRADOR"
	PapelMecanico  PapelUsuario = "MECANICO"
	PapelAtendente PapelUsuario = "ATENDENTE"
	PapelCliente   PapelUsuario = "CLIENTE"
)

// Valido indica se p é um dos papéis conhecidos.
func (p PapelUsuario) Valido() bool {
	switch p {
	case PapelAdmin, PapelMecanico, PapelAtendente, PapelCliente:
		return true
	default:
		return false
	}
}
