package query

type FilterInput struct {
	Field    string
	Operator string
	Value    string
}

type ParamsInput struct {
	Page      int
	Order     string
	Direction string
	Filters   []FilterInput
}
