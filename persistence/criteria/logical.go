package criteria

import (
	"encoding"
	"fmt"
)

// A LogicalOperator is a kind of operator used by [Criteria] to concatenate filters.
//
// If [And] is specified, all filters must result in a true statement to match items.
// On the other hand, if [Or] is specified, just one filter is required to result in a true statement
// to match items.
type LogicalOperator uint8

const (
	And LogicalOperator = iota
	Or
)

var (
	// compile-time assertions
	_ fmt.Stringer             = (*LogicalOperator)(nil)
	_ encoding.TextMarshaler   = (*LogicalOperator)(nil)
	_ encoding.TextUnmarshaler = (*LogicalOperator)(nil)

	_logicalStringValMap = map[string]LogicalOperator{
		"and": And,
		"or":  Or,
		"AND": And,
		"OR":  Or,
	}
	_logicalValStringMap = map[LogicalOperator]string{
		And: "and",
		Or:  "or",
	}
)

// NewLogicalOperator allocates a new [LogicalOperator] based on its string value.
func NewLogicalOperator(v string) LogicalOperator {
	return _logicalStringValMap[v]
}

func (s LogicalOperator) String() string {
	return _logicalValStringMap[s]
}

func (s LogicalOperator) MarshalText() (text []byte, err error) {
	return []byte(s.String()), nil
}

func (s *LogicalOperator) UnmarshalText(text []byte) error {
	*s = NewLogicalOperator(string(text))
	return nil
}
