package auth

// TipoUsuario distingue a fonte de identidade do usuário autenticado.
type TipoUsuario string

const (
	TipoInterno TipoUsuario = "interno"
	TipoCliente TipoUsuario = "cliente"
)

// PapelUsuario distingue o papel do usuário dentro do seu TipoUsuario.
// Usuários TipoInterno têm um dos papéis admin/mecanico/atendente;
// usuários TipoCliente têm sempre PapelCliente.
type PapelUsuario string

const (
	PapelAdmin     PapelUsuario = "admin"
	PapelMecanico  PapelUsuario = "mecanico"
	PapelAtendente PapelUsuario = "atendente"
	PapelCliente   PapelUsuario = "cliente"
)
