package servico

import (
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
)

// NewQueryBuilder define os campos, tipos e operadores públicos da consulta de
// serviços. Campos não declarados aqui não podem virar filtro ou ordenação SQL.
func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewDefaultBuilder(
		map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"nome": {
				Column:    "nome",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"descricao": {
				Column:    "descricao",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"preco_base": {
				Column:    "preco_base",
				Type:      mysqlquery.ValueTypeFloat64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"tempo_estimado_minutos": {
				Column:    "tempo_estimado_minutos",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"criado_por": {
				Column:    "criado_por",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"ativo": {
				Column:    "ativo",
				Type:      mysqlquery.ValueTypeBool,
				Sortable:  true,
				Operators: mysqlquery.BoolOperators,
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
