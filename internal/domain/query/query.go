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

// Filter contém um filtro já extraído da camada de transporte.
type Filter struct {
	Field    string
	Operator Operator
	Value    string
}

// Params contém os parâmetros comuns de uma consulta paginada.
type Params struct {
	Offset    int
	Limit     int
	Order     string
	Direction Direction
	Filters   []Filter
}

// Page representa uma página e os metadados da consulta que a produziu.
type Page[T any] struct {
	Items     []T
	Total     int64
	Offset    int
	Limit     int
	Order     string
	Direction Direction
}
