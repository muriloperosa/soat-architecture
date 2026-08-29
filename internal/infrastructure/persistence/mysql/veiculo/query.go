package veiculo

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

func NewQueryBuilder() *mysqlquery.Builder {
	return mysqlquery.NewBuilder(mysqlquery.Config{
		Fields: map[string]mysqlquery.Field{
			"id": {
				Column:    "id",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"placa": {
				Column:    "placa",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"marca": {
				Column:    "marca",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"modelo": {
				Column:    "modelo",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"quilometragem_atual": {
				Column:    "quilometragem_atual",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"ano": {
				Column:    "ano",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"cor": {
				Column:    "cor",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"criado_por": {
				Column:    "criado_por",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
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
			"data_cadastro": {
				Column:    "data_cadastro",
				Type:      mysqlquery.ValueTypeTime,
				Sortable:  true,
				Operators: dateOperators,
			},
		},
		DefaultOrder:     "id",
		DefaultDirection: domainquery.DirectionASC,
		DefaultLimit:     20,
		MaxLimit:         100,
	})
}
