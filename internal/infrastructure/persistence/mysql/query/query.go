package query

import domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"

const DefaultPageSize = 20

var (
	TextOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorLike,
		domainquery.OperatorNotLike,
		domainquery.OperatorIn,
		domainquery.OperatorNotIn,
	}

	NumberOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorGreaterThan,
		domainquery.OperatorGreaterOrEq,
		domainquery.OperatorLessThan,
		domainquery.OperatorLessOrEq,
		domainquery.OperatorIn,
		domainquery.OperatorNotIn,
	}

	DateOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
		domainquery.OperatorGreaterThan,
		domainquery.OperatorGreaterOrEq,
		domainquery.OperatorLessThan,
		domainquery.OperatorLessOrEq,
	}

	BoolOperators = []domainquery.Operator{
		domainquery.OperatorEqual,
		domainquery.OperatorNotEqual,
	}
)

func NewDefaultBuilder(
	fields map[string]Field,
	defaultOrder string,
) *Builder {
	return NewBuilder(Config{
		Fields:           fields,
		DefaultOrder:     defaultOrder,
		DefaultDirection: domainquery.DirectionASC,
		PageSize:         DefaultPageSize,
	})
}
