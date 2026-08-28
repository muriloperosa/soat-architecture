// Package httpquery converte query parameters HTTP nos contratos genéricos de
// paginação e filtros do domínio.
package httpquery

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// Parser interpreta offset, limit, order, direction e os demais parâmetros
// como filtros diretos por campo.
type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// Parse aceita filtros no formato campo=valor. O sufixo _not representa
// negação. O operador efetivo é escolhido na persistência conforme o tipo.
func (p *Parser) Parse(c *gin.Context) (query.Params, error) {
	params := query.Params{
		Order:     strings.TrimSpace(c.Query("order")),
		Direction: query.Direction(strings.ToUpper(strings.TrimSpace(c.Query("direction")))),
		Filters:   make([]query.Filter, 0),
	}

	var err error
	if rawOffset, ok := c.GetQuery("offset"); ok {
		params.Offset, err = parseInteger("offset", rawOffset)
		if err != nil {
			return query.Params{}, err
		}
	}

	if rawLimit, ok := c.GetQuery("limit"); ok {
		params.Limit, err = parseInteger("limit", rawLimit)
		if err != nil {
			return query.Params{}, err
		}
	}

	queryValues := c.Request.URL.Query()
	keys := make([]string, 0, len(queryValues))
	for key := range queryValues {
		if !isReservedParameter(key) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, key := range keys {
		field := strings.ToLower(strings.TrimSpace(key))
		operator := query.OperatorAuto
		if strings.HasSuffix(field, "_not") {
			field = strings.TrimSuffix(field, "_not")
			operator = query.OperatorAutoNot
		}
		if field == "" {
			return query.Params{}, shared.NewValidationError("Campo do filtro é obrigatório.")
		}

		for _, value := range queryValues[key] {
			value = strings.TrimSpace(value)
			if value == "" {
				return query.Params{}, shared.NewValidationError(
					fmt.Sprintf("Valor do filtro %q é obrigatório.", field),
				)
			}
			params.Filters = append(params.Filters, query.Filter{
				Field: field, Operator: operator, Value: value,
			})
		}
	}

	return params, nil
}

func parseInteger(name, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, shared.NewValidationError(fmt.Sprintf("Parâmetro %s deve ser um número inteiro.", name))
	}
	return parsed, nil
}

func isReservedParameter(name string) bool {
	switch strings.ToLower(name) {
	case "offset", "limit", "order", "direction":
		return true
	default:
		return false
	}
}
