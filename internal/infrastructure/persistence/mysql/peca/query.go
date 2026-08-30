package peca

import (
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
)

func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewDefaultBuilder(
		map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"codigo": {
				Column:    "codigo",
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
			"marca": {
				Column:    "marca",
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
			"preco": {
				Column:    "preco",
				Type:      mysqlquery.ValueTypeFloat64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"quantidade_em_estoque": {
				Column:    "quantidade_em_estoque",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"estoque_minimo": {
				Column:    "estoque_minimo",
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
