package sql

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

func parseValue(col string, v any) exp.Expression {
	switch castedVal := v.(type) {
	case time.Time:
		return goqu.C(col).Gt(goqu.L("?::timestamptz", castedVal))
	case DateTimeUTC:
		//return fmt.Sprintf("%s::timestamptz", castedVal.Format(time.RFC3339Nano))
		return nil
	default:
		return nil
	}
}
