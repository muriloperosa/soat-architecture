package peca

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
			"codigo": {
				Column:    "codigo",
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
			"marca": {
				Column:    "marca",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"descricao": {
				Column:    "descricao",
				Type:      mysqlquery.ValueTypeString,
				Sortable:  true,
				Operators: textOperators,
			},
			"preco": {
				Column:    "preco",
				Type:      mysqlquery.ValueTypeFloat64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"quantidade_em_estoque": {
				Column:    "quantidade_em_estoque",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
			},
			"estoque_minimo": {
				Column:    "estoque_minimo",
				Type:      mysqlquery.ValueTypeUint64,
				Sortable:  true,
				Operators: numberOperators,
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
