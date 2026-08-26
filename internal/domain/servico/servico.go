package servico

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

// Servico é um item do catálogo da oficina (troca de óleo, alinhamento,
// revisão etc.), com preço base e tempo estimado. Aggregate root.
type Servico struct {
	id              uint64
	nome            string
	descricao       string
	precoBase       float64
	tempoEstimado   shared.DuracaoEstimada
	criadoPor       uint64
	ativo           bool
	dataCadastro    time.Time
	dataAtualizacao time.Time
}

// NewServico cria um Serviço novo, ativo, com as invariantes validadas.
// criadoPor é o ID do usuário interno responsável pelo cadastro.
func NewServico(nome, descricao string, precoBase float64, tempoMinutos int, criadoPor uint64) (*Servico, error) {
	nome, descricao, tempoEstimado, err := validarDados(nome, descricao, precoBase, tempoMinutos, criadoPor)
	if err != nil {
		return nil, err
	}

	agora := time.Now()
	return &Servico{
		nome:            nome,
		descricao:       descricao,
		precoBase:       precoBase,
		tempoEstimado:   tempoEstimado,
		criadoPor:       criadoPor,
		ativo:           true,
		dataCadastro:    agora,
		dataAtualizacao: agora,
	}, nil
}

// RestaurarServico reidrata um Serviço a partir de dados já persistidos;
// não reaplica validação de negócio. Usado só pelo mapper de persistência
// (internal/infrastructure/persistence/mysql/servico).
func RestaurarServico(
	id uint64,
	nome, descricao string,
	precoBase float64,
	tempoEstimado shared.DuracaoEstimada,
	criadoPor uint64,
	ativo bool,
	dataCadastro, dataAtualizacao time.Time,
) *Servico {
	return &Servico{
		id:              id,
		nome:            nome,
		descricao:       descricao,
		precoBase:       precoBase,
		tempoEstimado:   tempoEstimado,
		criadoPor:       criadoPor,
		ativo:           ativo,
		dataCadastro:    dataCadastro,
		dataAtualizacao: dataAtualizacao,
	}
}

// Atualizar troca nome, descrição, preço base e tempo estimado.
// Não altera criadoPor nem ativo.
func (s *Servico) Atualizar(nome, descricao string, precoBase float64, tempoMinutos int) error {
	nome, descricao, tempoEstimado, err := validarDados(nome, descricao, precoBase, tempoMinutos, s.criadoPor)
	if err != nil {
		return err
	}

	s.nome = nome
	s.descricao = descricao
	s.precoBase = precoBase
	s.tempoEstimado = tempoEstimado
	s.dataAtualizacao = time.Now()
	return nil
}

// Ativar reabilita o serviço para uso em novas Ordens de Serviço.
func (s *Servico) Ativar() {
	s.ativo = true
	s.dataAtualizacao = time.Now()
}

// Inativar bloqueia o serviço de ser usado em novas Ordens de Serviço.
// Não remove o registro (soft delete).
func (s *Servico) Inativar() {
	s.ativo = false
	s.dataAtualizacao = time.Now()
}

// AtribuirID preenche o ID gerado pelo banco após o insert. Só o repositório
// de persistência chama isso.
func (s *Servico) AtribuirID(id uint64) {
	s.id = id
}

func (s *Servico) ID() uint64                            { return s.id }
func (s *Servico) Nome() string                          { return s.nome }
func (s *Servico) Descricao() string                     { return s.descricao }
func (s *Servico) PrecoBase() float64                    { return s.precoBase }
func (s *Servico) TempoEstimado() shared.DuracaoEstimada { return s.tempoEstimado }
func (s *Servico) CriadoPor() uint64                     { return s.criadoPor }
func (s *Servico) Ativo() bool                           { return s.ativo }
func (s *Servico) DataCadastro() time.Time               { return s.dataCadastro }
func (s *Servico) DataAtualizacao() time.Time            { return s.dataAtualizacao }

func validarDados(nome, descricao string, precoBase float64, tempoMinutos int, criadoPor uint64) (string, string, shared.DuracaoEstimada, error) {
	nome = texts.NormalizeSpaces(nome)
	descricao = texts.NormalizeSpaces(descricao)
	if nome == "" {
		return "", "", shared.DuracaoEstimada{}, ErrNomeObrigatorio
	}
	if descricao == "" {
		return "", "", shared.DuracaoEstimada{}, ErrDescricaoObrigatoria
	}
	if precoBase < 0 {
		return "", "", shared.DuracaoEstimada{}, ErrPrecoInvalido
	}
	if criadoPor == 0 {
		return "", "", shared.DuracaoEstimada{}, ErrCriadoPorObrigatorio
	}
	tempoEstimado, err := shared.NewDuracaoEstimada(tempoMinutos)
	if err != nil {
		return "", "", shared.DuracaoEstimada{}, err
	}
	return nome, descricao, tempoEstimado, nil
}
