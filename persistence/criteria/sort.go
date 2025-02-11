package criteria

import (
	"encoding"
	"fmt"
)

// SortOperator is the kind of sorting method to use during a data set querying operation.
type SortOperator uint8

// Sort is a sentinel structure used to specify the field and the method of sorting for
// a data set query operation.
type Sort struct {
	// Field sort results based on the values of this.
	Field string
	// Operator the sorting mechanism (e.g. ascending, descending).
	Operator SortOperator
}

const (
	// SortUnknown sort method is not known.
	SortUnknown = SortOperator(iota)
	// SortAscending ascending sorting method.
	SortAscending
	// SortDescending descending sorting method.
	SortDescending
)

var (
	// compile-time assertions
	_ fmt.Stringer             = (*SortOperator)(nil)
	_ encoding.TextMarshaler   = (*SortOperator)(nil)
	_ encoding.TextUnmarshaler = (*SortOperator)(nil)

	_sortStringValMap = map[string]SortOperator{
		"asc":  SortAscending,
		"desc": SortDescending,
		"ASC":  SortAscending,
		"DESC": SortDescending,
	}
	_sortValStringMap = map[SortOperator]string{
		SortAscending:  "asc",
		SortDescending: "desc",
	}
)

// NewSortOperator allocates a new [SortOperator] based on its string value.
func NewSortOperator(v string) SortOperator {
	return _sortStringValMap[v]
}

func (s SortOperator) String() string {
	return _sortValStringMap[s]
}

func (s SortOperator) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *SortOperator) UnmarshalText(text []byte) error {
	*s = NewSortOperator(string(text))
	return nil
}
