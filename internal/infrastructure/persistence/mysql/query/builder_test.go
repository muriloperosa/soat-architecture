package query

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type testModel struct {
	ID        uint64
	Name      string
	Active    bool
	CreatedAt string
}

func (testModel) TableName() string {
	return "test_models"
}

func testBuilder() *Builder {
	return NewBuilder(Config{
		Fields: map[string]Field{
			"id": {
				Column: "id", Type: ValueTypeUint64, Sortable: true,
				Operators: []domainquery.Operator{
					domainquery.OperatorEqual,
					domainquery.OperatorIn,
					domainquery.OperatorNotIn,
				},
			},
			"name": {
				Column: "name", Type: ValueTypeString, Sortable: true,
				Operators: []domainquery.Operator{
					domainquery.OperatorEqual,
					domainquery.OperatorLike,
					domainquery.OperatorNotLike,
					domainquery.OperatorIsNull,
				},
			},
			"active": {
				Column: "active", Type: ValueTypeBool,
				Operators: []domainquery.Operator{domainquery.OperatorEqual},
			},
			"created_at": {
				Column: "created_at", Type: ValueTypeTime,
				Operators: []domainquery.Operator{
					domainquery.OperatorEqual,
					domainquery.OperatorGreaterOrEq,
					domainquery.OperatorLessThan,
					domainquery.OperatorLessOrEq,
				},
			},
		},
		DefaultOrder:     "id",
		DefaultDirection: domainquery.DirectionASC,
		PageSize:         20,
	})
}

func TestBuilderResolveFiltrosAutomaticosPorTipo(t *testing.T) {
	builder := testBuilder()
	db, err := builder.ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{
		{Field: "name", Operator: domainquery.OperatorAuto, Value: "Caio"},
		{Field: "id", Operator: domainquery.OperatorAuto, Value: "1,2,3"},
		{Field: "active", Operator: domainquery.OperatorAuto, Value: "true"},
	})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "name LIKE ?")
	require.Contains(t, statement.SQL.String(), "id IN")
	require.Contains(t, statement.SQL.String(), "active = ?")
	require.Contains(t, statement.Vars, "%Caio%")
	require.Contains(t, statement.Vars, true)
}

func TestBuilderResolveNegacaoAutomaticaParaTexto(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "name", Operator: domainquery.OperatorAutoNot, Value: "Caio",
	}})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "name NOT LIKE ?")
	require.Contains(t, statement.Vars, "%Caio%")
}

func TestBuilderResolveIntervaloDeDatasISO8601(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAuto,
		Value: "2026-08-20,2026-08-22",
	}})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "created_at >= ?")
	require.Contains(t, statement.SQL.String(), "created_at < ?")
	require.Len(t, statement.Vars, 2)
}

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn: sqlDB, SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true})
	require.NoError(t, err)
	return db
}

func TestBuilderNormalizeAplicaPadroes(t *testing.T) {
	params, err := testBuilder().Normalize(domainquery.Params{})

	require.NoError(t, err)
	require.Equal(t, 1, params.Page)
	require.Equal(t, 20, testBuilder().PageSize())
	require.Equal(t, "id", params.Order)
	require.Equal(t, domainquery.DirectionASC, params.Direction)
}

func TestBuilderNormalizeRejeitaOrdenacaoNaoPermitida(t *testing.T) {
	_, err := testBuilder().Normalize(domainquery.Params{Order: "id; DROP TABLE clientes"})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderApplyFiltersEApplyPagination(t *testing.T) {
	builder := testBuilder()
	params, err := builder.Normalize(domainquery.Params{
		Page: 2, Order: "name", Direction: domainquery.DirectionDESC,
		Filters: []domainquery.Filter{
			{Field: "name", Operator: domainquery.OperatorLike, Value: "Maria"},
			{Field: "id", Operator: domainquery.OperatorIn, Value: "1|2|3"},
		},
	})
	require.NoError(t, err)

	db, err := builder.ApplyFilters(dryRunDB(t).Model(&testModel{}), params.Filters)
	require.NoError(t, err)
	statement := builder.ApplyPagination(db, params).Find(&[]testModel{}).Statement

	require.Contains(t, statement.SQL.String(), "name LIKE ?")
	require.Contains(t, statement.SQL.String(), "id IN")
	require.Contains(t, statement.SQL.String(), "ORDER BY name DESC")
	require.Contains(t, statement.SQL.String(), "LIMIT ? OFFSET ?")
	require.Contains(t, statement.Vars, "%Maria%")
}

func TestBuilderApplyFiltersRejeitaTipoInvalido(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "active", Operator: domainquery.OperatorEqual, Value: "talvez",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderApplyFiltersRejeitaOperadorNaoPermitido(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "active", Operator: domainquery.OperatorLike, Value: "true",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderTotalPages(t *testing.T) {
	builder := testBuilder()

	require.Zero(t, builder.TotalPages(0))
	require.Equal(t, 1, builder.TotalPages(1))
	require.Equal(t, 1, builder.TotalPages(20))
	require.Equal(t, 2, builder.TotalPages(21))
}

func TestNewDefaultBuilder(t *testing.T) {
	builder := NewDefaultBuilder(map[string]Field{
		"id": {Column: "id", Type: ValueTypeUint64, Sortable: true, Operators: []domainquery.Operator{domainquery.OperatorEqual}},
	}, "id")

	require.Equal(t, DefaultPageSize, builder.PageSize())

	params, err := builder.Normalize(domainquery.Params{})
	require.NoError(t, err)
	require.Equal(t, "id", params.Order)
	require.Equal(t, domainquery.DirectionASC, params.Direction)
}

func TestBuilderNormalizeRejeitaPageNegativo(t *testing.T) {
	_, err := testBuilder().Normalize(domainquery.Params{Page: -1})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderNormalizeRejeitaDirectionInvalida(t *testing.T) {
	_, err := testBuilder().Normalize(domainquery.Params{Order: "id", Direction: "SIDEWAYS"})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderResolveAutomaticoNumerosNegadoEMultiplo(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "id", Operator: domainquery.OperatorAutoNot, Value: "1,2",
	}})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "id NOT IN")
}

func TestBuilderResolveAutomaticoBooleanoRejeitaMultiplosValores(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "active", Operator: domainquery.OperatorAuto, Value: "true,false",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderResolveAutomaticoDataRejeitaNegacao(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAutoNot, Value: "2026-08-20",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderResolveAutomaticoDataRejeitaMaisDeDoisValores(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAuto, Value: "2026-08-20,2026-08-21,2026-08-22",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderResolveAutomaticoDataInicioAposFimRejeitado(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAuto, Value: "2026-08-22,2026-08-20",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderResolveAutomaticoDataUnicaComHorario(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAuto, Value: "2026-08-20T10:00:00Z",
	}})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "created_at = ?")
}

func TestBuilderResolveAutomaticoDataIntervaloComHorarioNoFim(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "created_at", Operator: domainquery.OperatorAuto, Value: "2026-08-20,2026-08-22T10:00:00Z",
	}})
	require.NoError(t, err)

	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "created_at <= ?")
}

func TestBuilderResolveAutomaticoRejeitaValorVazio(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "name", Operator: domainquery.OperatorAuto, Value: "a,,b",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestBuilderApplyFiltersIsNullENotIn(t *testing.T) {
	db, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{
		{Field: "name", Operator: domainquery.OperatorIsNull},
	})
	require.NoError(t, err)
	statement := db.Find(&[]testModel{}).Statement
	require.Contains(t, statement.SQL.String(), "name IS NULL")
}

func TestBuilderApplyFiltersRejeitaValorNaoNumerico(t *testing.T) {
	_, err := testBuilder().ApplyFilters(dryRunDB(t).Model(&testModel{}), []domainquery.Filter{{
		Field: "id", Operator: domainquery.OperatorEqual, Value: "abc",
	}})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}
