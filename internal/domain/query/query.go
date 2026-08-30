// Package query define os contratos de paginação, ordenação e filtros usados
// pelas portas do domínio, sem acoplar os casos de uso ao GORM ou ao HTTP.
package query

// Direction representa a direção da ordenação.
type Direction string

const (
	DirectionASC  Direction = "ASC"
	DirectionDESC Direction = "DESC"
)

// Operator representa uma operação de filtro suportada pela consulta.
type Operator string

const (
	OperatorAuto        Operator = "auto"
	OperatorAutoNot     Operator = "auto_not"
	OperatorEqual       Operator = "eq"
	OperatorNotEqual    Operator = "neq"
	OperatorLike        Operator = "like"
	OperatorNotLike     Operator = "not_like"
	OperatorGreaterThan Operator = "gt"
	OperatorGreaterOrEq Operator = "gte"
	OperatorLessThan    Operator = "lt"
	OperatorLessOrEq    Operator = "lte"
	OperatorIn          Operator = "in"
	OperatorNotIn       Operator = "not_in"
	OperatorIsNull      Operator = "is_null"
	OperatorIsNotNull   Operator = "is_not_null"
)

// Filter contém um filtro já extraído da camada de aplicação.
type Filter struct {
	Field    string
	Operator Operator
	Value    string
}

// Params contém os parâmetros usados pelas portas de consulta do domínio.
// Page é baseada em 1. O tamanho da página é uma decisão da persistência.
type Params struct {
	Page      int
	Order     string
	Direction Direction
	Filters   []Filter
}

// Page representa uma página e os metadados da consulta que a produziu.
type Page[T any] struct {
	Items      []T
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	Order      string
	Direction  Direction
}
