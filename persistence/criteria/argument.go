package criteria

import "github.com/hadroncorp/geck/persistence/paging/pagetoken"

// ArgumentTemplate is a sentinel structure used to specify certain fields of [Criteria].
//
// This structure contains validations tags and avoids leaking internal types
// to external clients.
//
// Embed this structure to other query structures to have general-purposed [Criteria] fields
// out-the-box.
type ArgumentTemplate struct {
	PageSize  int64 `validate:"omitempty,gte=1,lte=250"`
	PageToken *pagetoken.Token
	Sort      SortArgumentTemplate `validate:"omitempty,dive"`
}

// SortArgumentTemplate is a primitive-only structure based on [Sort].
//
// This structure contains validations tags and avoids leaking internal types
// to external clients.
type SortArgumentTemplate struct {
	Field    string `validate:"omitempty,lte=96"`
	Operator string `validate:"omitempty,oneof=ASC DESC"`
}

// ToSort converts [SortArgumentTemplate] to a [Sort].
func (q SortArgumentTemplate) ToSort() Sort {
	return Sort{
		Field:    q.Field,
		Operator: NewSortOperator(q.Operator),
	}
}
