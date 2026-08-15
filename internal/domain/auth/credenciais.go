package auth

import "context"

// RepositorioCredenciais busca a credencial de login por email.
// Uma interface, duas implementações: usuariointerno e cliente
// (injetadas no wiring por endpoint, não decididas em runtime).
type RepositorioCredenciais interface {
	BuscarPorEmail(ctx context.Context, email string) (*Credencial, error)
}

// Credencial é a projeção mínima necessária pro fluxo de login.
type Credencial struct {
	ID        string
	SenhaHash string
	Papel     PapelUsuario
}
