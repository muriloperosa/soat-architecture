package cliente

import (
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
)

var (
	textOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorLike,
		domainquery.OperatorNotLike,
		domainquery.OperatorIn,
		domainquery.OperatorNotIn,
	}
	numberOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorGreaterThan,
		domainquery.OperatorGreaterOrEq,
		domainquery.OperatorLessThan,
		domainquery.OperatorLessOrEq,
		domainquery.OperatorIn,
		domainquery.OperatorNotIn,
	}
	dateOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorGreaterThan,
		domainquery.OperatorGreaterOrEq,
		domainquery.OperatorLessThan,
		domainquery.OperatorLessOrEq,
	}
)

// NewQueryBuilder define os campos, tipos e operadores públicos da consulta de
// clientes. Campos não declarados aqui não podem virar filtro ou ordenação SQL.
func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewBuilder(mysqlquery.Config{
		Fields: map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"documento": {
				Column:    "documento",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"tipo": {
				Column:    "tipo",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"nome": {
				Column:    "nome",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"email": {
				Column:    "email",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"telefone": {
				Column:    "telefone",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"ativo": {
				Column:   "ativo",
				Type:     mysqlquery.ValueTypeBool,
				Sortable: true,
				Operators: []domainquery.Operator{
					domainquery.OperatorEqual,
					domainquery.OperatorNotEqual,
				},
			},
			"requer_alterar_senha": {
				Column:   "requer_alterar_senha",
				Type:     mysqlquery.ValueTypeBool,
				Sortable: true,
				Operators: []domainquery.Operator{
					domainquery.OperatorEqual,
					domainquery.OperatorNotEqual,
				},
			},
			"criado_por": {
				Column:    "criado_por",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"data_cadastro": {
				Column: "data_cadastro", Type: mysqlquery.ValueTypeTime, Sortable: true,
				Operators: dateOperators,
			},
			"data_atualizacao": {
				Column: "data_atualizacao", Type: mysqlquery.ValueTypeTime, Sortable: true,
				Operators: dateOperators,
			},
		},
		DefaultOrder:     "id",
		DefaultDirection: domainquery.DirectionASC,
		DefaultLimit:     20,
		MaxLimit:         100,
	})
}
