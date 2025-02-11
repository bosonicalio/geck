package criteria

import (
	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence/paging/pagetoken"
)

// Criteria is a structure containing fields used to indicate routines how to
// retrieve items from a data set in a more fine-grained way.
//
// For more information, read about the criteria pattern.
type Criteria struct {
	PageSize  int64
	PageToken *pagetoken.Token
	Sorting   Sort
	Operator  LogicalOperator
	Filters   []Filter
}

// HasSorting checks if [Criteria] has a valid sorting specification.
func (c Criteria) HasSorting() bool {
	return c.Sorting.Field != "" && c.Sorting.Operator != SortUnknown
}

// HasPreviousPageToken checks if [Criteria] has a previous page token.
func (c Criteria) HasPreviousPageToken() bool {
	if c.PageToken == nil {
		return false
	}
	return c.PageToken.Direction == pagetoken.PreviousDirection
}

// HasNextPageToken checks if [Criteria] has a next page token.
func (c Criteria) HasNextPageToken() bool {
	if c.PageToken == nil {
		return false
	}
	return c.PageToken.Direction == pagetoken.NextDirection
}

// -- OPTIONS --

// Option is a routine used to set several [Criteria] optional fields.
type Option func(*Criteria)

// WithSorting sets the sorting method of a [Criteria].
func WithSorting(field string, op SortOperator) Option {
	if field == "" || op == SortUnknown {
		// no-op
		return func(_ *Criteria) {}
	}

	return func(criteria *Criteria) {
		criteria.Sorting = Sort{
			Field:    field,
			Operator: op,
		}
	}
}

// WithSort sets [Sort] method of a [Criteria].
func WithSort(s Sort) Option {
	return func(criteria *Criteria) {
		criteria.Sorting = s
	}
}

// WithFilter appends a [Filter] to a [Criteria].
func WithFilter(field string, op FilterOperator, values ...any) Option {
	return func(criteria *Criteria) {
		if criteria.Filters == nil {
			criteria.Filters = make([]Filter, 0, 1)
		}
		criteria.Filters = append(criteria.Filters, Filter{
			Field:    field,
			Operator: op,
			Values:   values,
		})
	}
}

// WithEmptyableFilter appends a [Filter] to a [Criteria].
//
// If `values` contains purely zero-values, a no-op [Option] will be allocated instead.
func WithEmptyableFilter[T comparable](field string, op FilterOperator, values ...T) Option {
	var zeroValue T
	emptyValue := lo.CoalesceOrEmpty(values...)
	if emptyValue == zeroValue {
		// no-op
		return func(_ *Criteria) {}
	}

	return func(criteria *Criteria) {
		if criteria.Filters == nil {
			criteria.Filters = make([]Filter, 0, 1)
		}
		// only allow non-empty fil
		values = lo.Filter(values, func(item T, _ int) bool {
			return item != zeroValue
		})
		criteria.Filters = append(criteria.Filters, Filter{
			Field:    field,
			Operator: op,
			Values: lo.Map(values, func(item T, index int) any {
				return item
			}),
		})
	}
}

// WithPageSize sets the size of the result set for a [Criteria] operation.
func WithPageSize(size int64) Option {
	return func(criteria *Criteria) {
		criteria.PageSize = size
	}
}

// WithOperator sets the concatenation method of a [Criteria].
//
// Also known as logical operator ([LogicalOperator]).
func WithOperator(op LogicalOperator) Option {
	return func(criteria *Criteria) {
		criteria.Operator = op
	}
}

func WithPageToken(t *pagetoken.Token) Option {
	return func(criteria *Criteria) {
		criteria.PageToken = t
	}
}
