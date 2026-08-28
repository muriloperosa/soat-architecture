// Package query aplica paginação, ordenação e filtros tipados em consultas
// GORM. Campos e operadores precisam ser declarados previamente na configuração.
package query

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"gorm.io/gorm"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeUint64  ValueType = "uint64"
	ValueTypeFloat64 ValueType = "float64"
	ValueTypeBool    ValueType = "bool"
	ValueTypeTime    ValueType = "time"
)

// Field define o nome da coluna, o tipo do valor e as operações permitidas.
// Column nunca deve ser preenchido com dados vindos da requisição.
type Field struct {
	Column    string
	Type      ValueType
	Sortable  bool
	Operators []domainquery.Operator
}

type Config struct {
	Fields           map[string]Field
	DefaultOrder     string
	DefaultDirection domainquery.Direction
	DefaultLimit     int
	MaxLimit         int
}

type Builder struct {
	config    Config
	operators map[string]map[domainquery.Operator]struct{}
}

func NewBuilder(config Config) *Builder {
	operators := make(map[string]map[domainquery.Operator]struct{}, len(config.Fields))
	for name, field := range config.Fields {
		operators[name] = make(map[domainquery.Operator]struct{}, len(field.Operators))
		for _, operator := range field.Operators {
			operators[name][operator] = struct{}{}
		}
	}

	return &Builder{config: config, operators: operators}
}

// Normalize aplica os padrões e valida os parâmetros que não dependem do tipo
// de cada filtro.
func (b *Builder) Normalize(params domainquery.Params) (domainquery.Params, error) {
	if params.Offset < 0 {
		return domainquery.Params{}, shared.NewValidationError("Offset não pode ser negativo.")
	}
	if params.Limit < 0 {
		return domainquery.Params{}, shared.NewValidationError("Limit não pode ser negativo.")
	}
	if params.Limit == 0 {
		params.Limit = b.config.DefaultLimit
	}
	if params.Limit > b.config.MaxLimit {
		return domainquery.Params{}, shared.NewValidationError(
			fmt.Sprintf("Limit não pode ser maior que %d.", b.config.MaxLimit),
		)
	}

	if params.Order == "" {
		params.Order = b.config.DefaultOrder
	}
	field, ok := b.config.Fields[params.Order]
	if !ok || !field.Sortable {
		return domainquery.Params{}, shared.NewValidationError("Campo de ordenação inválido.")
	}

	if params.Direction == "" {
		params.Direction = b.config.DefaultDirection
	}
	params.Direction = domainquery.Direction(strings.ToUpper(string(params.Direction)))
	if params.Direction != domainquery.DirectionASC && params.Direction != domainquery.DirectionDESC {
		return domainquery.Params{}, shared.NewValidationError("Direction deve ser ASC ou DESC.")
	}

	return params, nil
}

// ApplyFilters aplica somente filtros com campo e operador permitidos. Todos os
// valores são enviados ao GORM como parâmetros, nunca concatenados ao SQL.
func (b *Builder) ApplyFilters(db *gorm.DB, filters []domainquery.Filter) (*gorm.DB, error) {
	for _, filter := range filters {
		field, ok := b.config.Fields[filter.Field]
		if !ok {
			return nil, shared.NewValidationError(
				fmt.Sprintf("Filtro não permitido para o campo %q.", filter.Field),
			)
		}
		resolved := []domainquery.Filter{filter}
		if filter.Operator == domainquery.OperatorAuto || filter.Operator == domainquery.OperatorAutoNot {
			automaticFilters, resolveErr := resolveAutomaticFilter(field, filter)
			if resolveErr != nil {
				return nil, resolveErr
			}
			resolved = automaticFilters
		}

		for _, resolvedFilter := range resolved {
			if _, ok = b.operators[filter.Field][resolvedFilter.Operator]; !ok {
				return nil, shared.NewValidationError(
					fmt.Sprintf(
						"Operador %q não permitido para o campo %q.",
						resolvedFilter.Operator,
						filter.Field,
					),
				)
			}

			filteredDB, applyErr := applyFilter(db, field, resolvedFilter)
			if applyErr != nil {
				return nil, applyErr
			}
			db = filteredDB
		}
	}

	return db, nil
}

func resolveAutomaticFilter(field Field, filter domainquery.Filter) ([]domainquery.Filter, error) {
	values, err := splitAutomaticValues(filter.Value)
	if err != nil {
		return nil, invalidValue(filter, err)
	}
	negated := filter.Operator == domainquery.OperatorAutoNot

	switch field.Type {
	case ValueTypeString:
		operator := domainquery.OperatorLike
		if negated {
			operator = domainquery.OperatorNotLike
		}
		if len(values) > 1 {
			operator = domainquery.OperatorIn
			if negated {
				operator = domainquery.OperatorNotIn
			}
		}
		return []domainquery.Filter{{
			Field: filter.Field, Operator: operator, Value: strings.Join(values, "|"),
		}}, nil
	case ValueTypeUint64, ValueTypeFloat64:
		operator := domainquery.OperatorEqual
		if negated {
			operator = domainquery.OperatorNotEqual
		}

		if len(values) > 1 {
			operator = domainquery.OperatorIn
			if negated {
				operator = domainquery.OperatorNotIn
			}
		}

		return []domainquery.Filter{{
			Field:    filter.Field,
			Operator: operator,
			Value:    strings.Join(values, "|"),
		}}, nil
	case ValueTypeBool:
		if len(values) != 1 {
			return nil, shared.NewValidationError(
				fmt.Sprintf("Filtro booleano %q aceita somente um valor.", filter.Field),
			)
		}
		operator := domainquery.OperatorEqual
		if negated {
			operator = domainquery.OperatorNotEqual
		}
		return []domainquery.Filter{{
			Field: filter.Field, Operator: operator, Value: values[0],
		}}, nil
	case ValueTypeTime:
		if negated {
			return nil, shared.NewValidationError(
				fmt.Sprintf("Negação automática não é suportada para o campo de data %q.", filter.Field),
			)
		}
		return resolveAutomaticTimeFilter(filter, values)
	default:
		return nil, shared.NewValidationError(
			fmt.Sprintf("Tipo do filtro %q não suporta operação automática.", filter.Field),
		)
	}
}

func resolveAutomaticTimeFilter(
	filter domainquery.Filter,
	values []string,
) ([]domainquery.Filter, error) {
	if len(values) > 2 {
		return nil, shared.NewValidationError(
			fmt.Sprintf("Filtro de data %q aceita uma data ou um intervalo.", filter.Field),
		)
	}

	start, startDateOnly, err := parseTimeValue(values[0])
	if err != nil {
		return nil, invalidValue(filter, err)
	}

	if len(values) == 1 {
		if !startDateOnly {
			return []domainquery.Filter{{
				Field: filter.Field, Operator: domainquery.OperatorEqual, Value: values[0],
			}}, nil
		}
		return []domainquery.Filter{
			{
				Field: filter.Field, Operator: domainquery.OperatorGreaterOrEq,
				Value: start.Format(time.RFC3339),
			},
			{
				Field: filter.Field, Operator: domainquery.OperatorLessThan,
				Value: start.AddDate(0, 0, 1).Format(time.RFC3339),
			},
		}, nil
	}

	end, endDateOnly, err := parseTimeValue(values[1])
	if err != nil {
		return nil, invalidValue(filter, err)
	}
	if start.After(end) {
		return nil, shared.NewValidationError(
			fmt.Sprintf("Data inicial do filtro %q deve ser anterior à data final.", filter.Field),
		)
	}

	endOperator := domainquery.OperatorLessOrEq
	if endDateOnly {
		end = end.AddDate(0, 0, 1)
		endOperator = domainquery.OperatorLessThan
	}

	return []domainquery.Filter{
		{
			Field: filter.Field, Operator: domainquery.OperatorGreaterOrEq,
			Value: start.Format(time.RFC3339),
		},
		{
			Field: filter.Field, Operator: endOperator,
			Value: end.Format(time.RFC3339),
		},
	}, nil
}

func splitAutomaticValues(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, fmt.Errorf("lista contém valor vazio")
		}
		values = append(values, value)
	}
	return values, nil
}

func parseTimeValue(raw string) (time.Time, bool, error) {
	if parsed, err := time.Parse(time.DateOnly, raw); err == nil {
		return parsed, true, nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	return parsed, false, err
}

// ApplyPagination aplica ordenação e janela após a contagem total.
func (b *Builder) ApplyPagination(db *gorm.DB, params domainquery.Params) *gorm.DB {
	field := b.config.Fields[params.Order]
	return db.Order(field.Column + " " + string(params.Direction)).Offset(params.Offset).Limit(params.Limit)
}

func applyFilter(db *gorm.DB, field Field, filter domainquery.Filter) (*gorm.DB, error) {
	column := field.Column

	switch filter.Operator {
	case domainquery.OperatorIsNull:
		return db.Where(column + " IS NULL"), nil
	case domainquery.OperatorIsNotNull:
		return db.Where(column + " IS NOT NULL"), nil
	case domainquery.OperatorIn, domainquery.OperatorNotIn:
		values, err := parseList(field.Type, filter.Value)
		if err != nil {
			return nil, invalidValue(filter, err)
		}
		operation := " IN ?"
		if filter.Operator == domainquery.OperatorNotIn {
			operation = " NOT IN ?"
		}
		return db.Where(column+operation, values), nil
	case domainquery.OperatorLike, domainquery.OperatorNotLike:
		value, err := parseValue(field.Type, filter.Value)
		if err != nil {
			return nil, invalidValue(filter, err)
		}
		text, ok := value.(string)
		if !ok {
			return nil, shared.NewValidationError(
				fmt.Sprintf("Operador %q exige um campo textual.", filter.Operator),
			)
		}
		operation := " LIKE ?"
		if filter.Operator == domainquery.OperatorNotLike {
			operation = " NOT LIKE ?"
		}
		return db.Where(column+operation, "%"+text+"%"), nil
	default:
		value, err := parseValue(field.Type, filter.Value)
		if err != nil {
			return nil, invalidValue(filter, err)
		}
		operation := map[domainquery.Operator]string{
			domainquery.OperatorEqual:       " = ?",
			domainquery.OperatorNotEqual:    " <> ?",
			domainquery.OperatorGreaterThan: " > ?",
			domainquery.OperatorGreaterOrEq: " >= ?",
			domainquery.OperatorLessThan:    " < ?",
			domainquery.OperatorLessOrEq:    " <= ?",
		}[filter.Operator]
		if operation == "" {
			return nil, shared.NewValidationError(fmt.Sprintf("Operador %q inválido.", filter.Operator))
		}
		return db.Where(column+operation, value), nil
	}
}

func parseList(valueType ValueType, raw string) ([]any, error) {
	parts := strings.FieldsFunc(raw, func(character rune) bool {
		return character == '|' || character == ','
	})
	values := make([]any, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return nil, fmt.Errorf("lista contém valor vazio")
		}
		value, err := parseValue(valueType, strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func parseValue(valueType ValueType, raw string) (any, error) {
	switch valueType {
	case ValueTypeString:
		return raw, nil

	case ValueTypeUint64:
		return strconv.ParseUint(raw, 10, 64)

	case ValueTypeFloat64:
		return strconv.ParseFloat(raw, 64)

	case ValueTypeBool:
		return strconv.ParseBool(raw)

	case ValueTypeTime:
		if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
			return parsed, nil
		}

		return time.Parse(time.DateOnly, raw)

	default:
		return nil, fmt.Errorf("tipo de filtro não suportado")
	}
}

func invalidValue(filter domainquery.Filter, err error) error {
	return shared.NewValidationError(
		fmt.Sprintf("Valor inválido para o filtro %q: %v.", filter.Field, err),
	)
}
