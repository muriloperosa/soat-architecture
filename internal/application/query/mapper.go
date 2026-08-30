package query

import domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"

func ToDomainParams(input ParamsInput) domainquery.Params {
	var filters []domainquery.Filter

	if len(input.Filters) > 0 {
		filters = make([]domainquery.Filter, 0, len(input.Filters))

		for _, filter := range input.Filters {
			filters = append(filters, domainquery.Filter{
				Field:    filter.Field,
				Operator: domainquery.Operator(filter.Operator),
				Value:    filter.Value,
			})
		}
	}

	return domainquery.Params{
		Page:      input.Page,
		Order:     input.Order,
		Direction: domainquery.Direction(input.Direction),
		Filters:   filters,
	}
}
