package query

import (
	"testing"

	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/stretchr/testify/require"
)

func TestToDomainParams(t *testing.T) {
	input := ParamsInput{
		Page: 2, Order: "nome", Direction: "desc",
		Filters: []FilterInput{{Field: "status", Operator: "eq", Value: "ativo"}},
	}

	got := ToDomainParams(input)

	require.Equal(t, 2, got.Page)
	require.Equal(t, "nome", got.Order)
	require.Equal(t, domainquery.Direction("desc"), got.Direction)
	require.Equal(t, []domainquery.Filter{{Field: "status", Operator: domainquery.Operator("eq"), Value: "ativo"}}, got.Filters)
}

func TestToDomainParamsSemFiltros(t *testing.T) {
	got := ToDomainParams(ParamsInput{Page: 1})
	require.Nil(t, got.Filters)
}
