package auth

// TipoUsuario distingue a fonte de identidade do usuário autenticado.
type TipoUsuario string

const (
	TipoInterno TipoUsuario = "interno"
	TipoCliente TipoUsuario = "cliente"
)
