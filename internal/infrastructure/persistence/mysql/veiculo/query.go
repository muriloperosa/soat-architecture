package veiculo

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
			"placa": {
				Column:    "placa",
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
			"modelo": {
				Column:    "modelo",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"quilometragem_atual": {
				Column:    "quilometragem_atual",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"ano": {
				Column:    "ano",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"cor": {
				Column:    "cor",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
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
