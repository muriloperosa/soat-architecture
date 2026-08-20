package veiculo

import (
	"time"
)

const anoMinimo = 1900

type Veiculo struct {
	id                 uint64
	placa              Placa
	marca              string
	modelo             string
	quilometragemAtual uint32
	ano                uint16
	cor                Cor
	criadoPor          uint64
	ativo              bool
	dataCadastro       time.Time
	dataAtualizacao    time.Time
}

func NewVeiculo(placa, marca, modelo string, quilometragemAtual uint32, ano uint16, cor string, criadoPor uint64) (*Veiculo, error) {
	placaVO, err := NewPlaca(placa)

	if err != nil {
		return nil, err
	}

	if marca == "" {
		return nil, ErrMarcaObrigatoria
	}

	if modelo == "" {
		return nil, ErrModeloObrigatorio
	}

	corVO, err := NewCor(cor)

	if err != nil {
		return nil, err
	}

	if !anoValido(ano) {
		return nil, ErrAnoInvalido
	}

	if criadoPor == 0 {
		return nil, ErrCriadoPorObrigatorio
	}

	agora := time.Now()

	return &Veiculo{
		placa:              placaVO,
		marca:              marca,
		modelo:             modelo,
		quilometragemAtual: quilometragemAtual,
		ano:                ano,
		cor:                corVO,
		criadoPor:          criadoPor,
		ativo:              true,
		dataCadastro:       agora,
		dataAtualizacao:    agora,
	}, nil
}

func RestaurarVeiculo(id uint64, placa Placa, marca, modelo string, quilometragemAtual uint32, ano uint16, cor Cor, criadoPor uint64, ativo bool, dataCadastro, dataAtualizacao time.Time) *Veiculo {
	return &Veiculo{
		id:                 id,
		placa:              placa,
		marca:              marca,
		modelo:             modelo,
		quilometragemAtual: quilometragemAtual,
		ano:                ano,
		cor:                cor,
		criadoPor:          criadoPor,
		ativo:              ativo,
		dataCadastro:       dataCadastro,
		dataAtualizacao:    dataAtualizacao,
	}
}

func (veiculo *Veiculo) Atualizar(marca, modelo, cor string) error {
	if marca == "" {
		return ErrMarcaObrigatoria
	}

	if modelo == "" {
		return ErrModeloObrigatorio
	}

	corVO, err := NewCor(cor)

	if err != nil {
		return err
	}

	veiculo.marca = marca
	veiculo.modelo = modelo
	veiculo.cor = corVO
	veiculo.dataAtualizacao = time.Now()

	return nil
}

func (veiculo *Veiculo) AtualizarQuilometragem(quilometragem uint32) error {
	if quilometragem < veiculo.quilometragemAtual {
		return ErrQuilometragemInvalida
	}

	veiculo.quilometragemAtual = quilometragem
	veiculo.dataAtualizacao = time.Now()

	return nil
}

func (veiculo *Veiculo) Ativar() {
	veiculo.ativo = true
	veiculo.dataAtualizacao = time.Now()
}

func (veiculo *Veiculo) Inativar() {
	veiculo.ativo = false
	veiculo.dataAtualizacao = time.Now()
}

func (veiculo *Veiculo) AtribuirID(id uint64) {
	veiculo.id = id
}

func anoValido(ano uint16) bool {
	anoMaximo := uint16(time.Now().Year() + 1)

	return ano >= anoMinimo && ano <= anoMaximo
}

func (veiculo *Veiculo) ID() uint64                 { return veiculo.id }
func (veiculo *Veiculo) Placa() Placa               { return veiculo.placa }
func (veiculo *Veiculo) Marca() string              { return veiculo.marca }
func (veiculo *Veiculo) Modelo() string             { return veiculo.modelo }
func (veiculo *Veiculo) QuilometragemAtual() uint32 { return veiculo.quilometragemAtual }
func (veiculo *Veiculo) Ano() uint16                { return veiculo.ano }
func (veiculo *Veiculo) Cor() Cor                   { return veiculo.cor }
func (veiculo *Veiculo) CriadoPor() uint64          { return veiculo.criadoPor }
func (veiculo *Veiculo) Ativo() bool                { return veiculo.ativo }
func (veiculo *Veiculo) DataCadastro() time.Time    { return veiculo.dataCadastro }
func (veiculo *Veiculo) DataAtualizacao() time.Time { return veiculo.dataAtualizacao }
