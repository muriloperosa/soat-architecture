package usuario

import appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"

// toCriarInput converte o DTO HTTP de criação pro DTO de entrada do
// CriarUsuarioUseCase.
func toCriarInput(req CriarUsuarioRequest) appusuario.CriarUsuarioInput {
	return appusuario.CriarUsuarioInput{Nome: req.Nome, Email: req.Email, SenhaInicial: req.Senha, Papel: req.Papel}
}

// toAtualizarInput converte o DTO HTTP de atualização pro DTO de entrada do
// AtualizarUsuarioUseCase. id vem do path param, não do corpo da requisição.
func toAtualizarInput(id uint64, req AtualizarUsuarioRequest) appusuario.AtualizarUsuarioInput {
	return appusuario.AtualizarUsuarioInput{ID: id, Nome: req.Nome, Email: req.Email, SenhaNova: req.SenhaNova, Papel: req.Papel}
}

// toAlterarSenhaInput converte o DTO HTTP de troca de senha pro DTO de
// entrada do AlterarSenhaUseCase. usuarioID vem do subject do JWT (self-service),
// não do corpo da requisição.
func toAlterarSenhaInput(usuarioID uint64, req AlterarSenhaRequest) appusuario.AlterarSenhaInput {
	return appusuario.AlterarSenhaInput{UsuarioID: usuarioID, SenhaNova: req.SenhaNova}
}

// toUsuarioResponse converte o DTO de saída dos use cases de gestão/consulta
// pra resposta HTTP comum (criação/atualização/consulta de usuário).
func toUsuarioResponse(out appusuario.UsuarioOutput) UsuarioResponse {
	return UsuarioResponse{
		ID:                 out.ID,
		Nome:               out.Nome,
		Email:              out.Email,
		Papel:              string(out.Papel),
		Ativo:              out.Ativo,
		RequerAlterarSenha: out.RequerAlterarSenha,
	}
}
