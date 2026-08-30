// Package httpquery interpreta os parâmetros de consulta HTTP sem expor
// contratos de domínio ou de persistência aos handlers.
package httpquery

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

const (
	OperatorAuto    = "auto"
	OperatorAutoNot = "auto_not"
)

// Filter representa um filtro extraído da query string HTTP.
type Filter struct {
	Field    string
	Operator string
	Value    string
}

// Params representa somente os parâmetros expostos pelo contrato HTTP.
type Params struct {
	Page      int
	Order     string
	Direction string
	Filters   []Filter
}

type Parser struct{}

func NewParser() *Parser { return &Parser{} }

// Parse aceita page, order e direction. Os demais parâmetros são filtros no
// formato campo=valor; o sufixo _not representa negação.
func (p *Parser) Parse(c *gin.Context) (Params, error) {
	params := Params{
		Page:      1,
		Order:     strings.TrimSpace(c.Query("order")),
		Direction: strings.ToUpper(strings.TrimSpace(c.Query("direction"))),
		Filters:   make([]Filter, 0),
	}

	if rawPage, ok := c.GetQuery("page"); ok {
		page, err := parseInteger("page", rawPage)
		if err != nil {
			return Params{}, err
		}
		if page < 1 {
			return Params{}, shared.NewValidationError("Parâmetro page deve ser maior ou igual a 1.")
		}
		params.Page = page
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
		operator := OperatorAuto
		if strings.HasSuffix(field, "_not") {
			field = strings.TrimSuffix(field, "_not")
			operator = OperatorAutoNot
		}
		if field == "" {
			return Params{}, shared.NewValidationError("Campo do filtro é obrigatório.")
		}

		for _, value := range queryValues[key] {
			value = strings.TrimSpace(value)
			if value == "" {
				return Params{}, shared.NewValidationError(
					fmt.Sprintf("Valor do filtro %q é obrigatório.", field),
				)
			}
			params.Filters = append(params.Filters, Filter{Field: field, Operator: operator, Value: value})
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
	case "page", "order", "direction":
		return true
	default:
		return false
	}
}
