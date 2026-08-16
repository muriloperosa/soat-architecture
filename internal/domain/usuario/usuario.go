package usuario

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// Usuario é o usuário interno (admin/mecânico/atendente). Fonte de
// autenticação JWT das APIs administrativas.
type Usuario struct {
	id                 uint64
	papel              shared.PapelUsuario
	nome               string
	email              shared.Email
	senha              shared.SenhaHash
	requerAlterarSenha bool
	ativo              bool
	dataCadastro       time.Time
	dataAtualizacao    time.Time
}

// NewUsuario cria um Usuario novo. A senha inicial (definida pelo
// administrador) nasce como provisória. AlterarSenha é obrigatório no
// primeiro acesso, forçado pela flag RequerAlterarSenha.
func NewUsuario(nome, email, senhaInicial string, papel shared.PapelUsuario) (*Usuario, error) {
	if nome == "" {
		return nil, ErrNomeObrigatorio
	}
	if !papel.Valido() {
		return nil, ErrPapelInvalido
	}
	emailVO, err := shared.NewEmail(email)
	if err != nil {
		return nil, err
	}
	senhaVO, err := shared.NewSenhaHash(senhaInicial)
	if err != nil {
		return nil, err
	}

	agora := time.Now()
	return &Usuario{
		papel:              papel,
		nome:               nome,
		email:              emailVO,
		senha:              senhaVO,
		requerAlterarSenha: true,
		ativo:              true,
		dataCadastro:       agora,
		dataAtualizacao:    agora,
	}, nil
}

// RestaurarUsuario reidrata um Usuario a partir de dados já persistidos;
// não reaplica validação de negócio nem regera a senha. Usado só pelo
// mapper de persistência (internal/infrastructure/persistence/mysql/usuario).
func RestaurarUsuario(id uint64, nome string, email shared.Email, senha shared.SenhaHash, papel shared.PapelUsuario, requerAlterarSenha, ativo bool, dataCadastro, dataAtualizacao time.Time) *Usuario {
	return &Usuario{
		id:                 id,
		papel:              papel,
		nome:               nome,
		email:              email,
		senha:              senha,
		requerAlterarSenha: requerAlterarSenha,
		ativo:              ativo,
		dataCadastro:       dataCadastro,
		dataAtualizacao:    dataAtualizacao,
	}
}

// AlterarSenha troca a senha e encerra o estado provisório; destrava o
// primeiro acesso.
func (u *Usuario) AlterarSenha(senhaNova string) error {
	senhaVO, err := shared.NewSenhaHash(senhaNova)
	if err != nil {
		return err
	}
	u.senha = senhaVO
	u.requerAlterarSenha = false
	u.dataAtualizacao = time.Now()
	return nil
}

// Atualizar troca nome e papel.
func (u *Usuario) Atualizar(nome string, papel shared.PapelUsuario) error {
	if nome == "" {
		return ErrNomeObrigatorio
	}
	if !papel.Valido() {
		return ErrPapelInvalido
	}
	u.nome = nome
	u.papel = papel
	u.dataAtualizacao = time.Now()
	return nil
}

// Ativar reabilita o usuário pra login.
func (u *Usuario) Ativar() {
	u.ativo = true
	u.dataAtualizacao = time.Now()
}

// Inativar bloqueia o usuário de logar (LoginUseCase rejeita via Credencial.Ativo).
func (u *Usuario) Inativar() {
	u.ativo = false
	u.dataAtualizacao = time.Now()
}

// AtribuirID preenche o ID gerado pelo banco após o insert. Só o repositório
// de persistência chama isso.
func (u *Usuario) AtribuirID(id uint64) {
	u.id = id
}

func (u *Usuario) ID() uint64                 { return u.id }
func (u *Usuario) Papel() shared.PapelUsuario { return u.papel }
func (u *Usuario) Nome() string               { return u.nome }
func (u *Usuario) Email() shared.Email        { return u.email }
func (u *Usuario) Senha() shared.SenhaHash    { return u.senha }
func (u *Usuario) RequerAlterarSenha() bool   { return u.requerAlterarSenha }
func (u *Usuario) Ativo() bool                { return u.ativo }
func (u *Usuario) DataCadastro() time.Time    { return u.dataCadastro }
func (u *Usuario) DataAtualizacao() time.Time { return u.dataAtualizacao }
