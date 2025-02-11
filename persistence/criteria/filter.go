package criteria

import (
	"encoding"
	"fmt"
)

// FilterOperator is the kind of filtering operation to be performed.
type FilterOperator uint16

// Filter is a structure used to select specific items from a data set.
type Filter struct {
	// Field indicates the filter to use values with this field name
	// to perform the filter operation.
	Field string
	// Operator indicates which filter algorithm will be used.
	Operator FilterOperator
	// Values slice of anonymous-type values passed to the filtering operation.
	Values []any
}

const (
	// FilterUnknown the filter is not known.
	FilterUnknown = FilterOperator(iota)
	Equal
	NotEqual
	GreaterThan
	LessThan
	GreaterThanOrEqualTo
	LessThanOrEqualTo
	In
	NotIn
	Like
	ILike
	NotLike
	NotILike
	Between
	NotBetween
	IsNil
	IsNotNil
)

var (
	// compile-time assertions
	_ fmt.Stringer             = (*FilterOperator)(nil)
	_ encoding.TextMarshaler   = (*FilterOperator)(nil)
	_ encoding.TextUnmarshaler = (*FilterOperator)(nil)

	_filterStringValMap = map[string]FilterOperator{
		"=":           Equal,
		"!=":          NotEqual,
		">":           GreaterThan,
		"<":           LessThan,
		">=":          GreaterThanOrEqualTo,
		"<=":          LessThanOrEqualTo,
		"in":          In,
		"not in":      NotIn,
		"like":        Like,
		"not like":    ILike,
		"ilike":       ILike,
		"not ilike":   ILike,
		"between":     Between,
		"not between": Between,
		"is nil":      IsNil,
		"is not nil":  IsNotNil,
	}
	_filterValStringMap = map[FilterOperator]string{
		Equal:                "=",
		NotEqual:             "!=",
		GreaterThan:          ">",
		LessThan:             "<",
		GreaterThanOrEqualTo: ">",
		LessThanOrEqualTo:    "<",
		In:                   "in",
		NotIn:                "not in",
		Like:                 "like",
		NotLike:              "not like",
		NotILike:             "not ilike",
		Between:              "between",
		NotBetween:           "not between",
		IsNil:                "is nil",
		IsNotNil:             "is not nil",
	}
)

// NewFilterOperator allocates a new [FilterOperator] based on its string value.
func NewFilterOperator(v string) FilterOperator {
	return _filterStringValMap[v]
}

func (s FilterOperator) String() string {
	return _filterValStringMap[s]
}

func (s FilterOperator) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *FilterOperator) UnmarshalText(text []byte) error {
	*s = NewFilterOperator(string(text))
	return nil
}
