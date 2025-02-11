package criteria

import "github.com/hadroncorp/geck/persistence/paging/pagetoken"

// Query is a sentinel structure used to specify certain fields of [Criteria].
//
// This structure contains validations tags and avoids leaking internal types
// to external clients.
//
// Embed this structure to other query structures to have general-purposed [Criteria] fields
// out-the-box.
type Query struct {
	PageSize  int64 `validate:"omitempty,gte=1,lte=250"`
	PageToken *pagetoken.Token
	Sort      SortQuery `validate:"omitempty,dive"`
}

// SortQuery is a primitive-only structure based on [Sort].
//
// This structure contains validations tags and avoids leaking internal types
// to external clients.
type SortQuery struct {
	Field    string `validate:"omitempty,lte=96"`
	Operator string `validate:"omitempty,oneof=ASC DESC"`
}

// ToSort converts [SortQuery] to a [Sort].
func (q SortQuery) ToSort() Sort {
	return Sort{
		Field:    q.Field,
		Operator: NewSortOperator(q.Operator),
	}
}
