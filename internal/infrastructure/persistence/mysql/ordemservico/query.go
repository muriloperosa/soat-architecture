package ordemservico

import mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"

// NewQueryBuilder define os campos, tipos e operadores públicos da consulta de
// Ordens de Serviço. Campos não declarados aqui não podem virar filtro ou ordenação SQL.
func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewDefaultBuilder(
		map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"numero": {
				Column:    "numero",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: mysqlquery.TextOperators,
			},
			"cliente_id": {
				Column:    "cliente_id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"veiculo_id": {
				Column:    "veiculo_id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"quilometragem_entrada": {
				Column:    "quilometragem_entrada",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: mysqlquery.NumberOperators,
			},
			"status": {
				Column:    "status",
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
			"data_cadastro": {
				Column:    "data_cadastro",
				Type:      mysqlquery.ValueTypeTime,
				Sortable:  true,
				Operators: mysqlquery.DateOperators,
			},
			"data_atualizacao": {
				Column:    "data_atualizacao",
				Type:      mysqlquery.ValueTypeTime,
				Sortable:  true,
				Operators: mysqlquery.DateOperators,
			},
		},
		"id",
	)
}
