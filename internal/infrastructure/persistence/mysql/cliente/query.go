package cliente

import (
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
)

// NewQueryBuilder define os campos, tipos e operadores públicos da consulta de
// clientes. Campos não declarados aqui não podem virar filtro ou ordenação SQL.
func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewDefaultBuilder(
		map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"documento": {
				Column:    "documento",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"tipo": {
				Column:    "tipo",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"nome": {
				Column:    "nome",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"email": {
				Column:    "email",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"telefone": {
				Column:    "telefone",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"ativo": {
				Column:    "ativo",
				Type:      mysqlquery.ValueTypeBool,
				Sortable:  true,
				Operators: mysqlquery.BoolOperators,
			},
			"criado_por": {
				Column:    "criado_por",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"data_cadastro": {
				Column:    "data_cadastro",
				Type:      mysqlquery.ValueTypeTime,
				Sortable:  true,
				Operators: mysqlquery.DateOperators,
			},
		},
		"id",
	)
}
