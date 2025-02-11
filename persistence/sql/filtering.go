package sql

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"

	"github.com/hadroncorp/geck/persistence/criteria"
)

func newFilterQuery(cr criteria.Criteria) (exp.ExpressionList, error) {
	expressions := make([]goqu.Expression, 0, len(cr.Filters)+1)
	for _, filter := range cr.Filters {
		if filter.Operator == criteria.IsNil {
			expressions = append(expressions, goqu.C(filter.Field).IsNull())
			continue
		} else if filter.Operator == criteria.IsNotNil {
			expressions = append(expressions, goqu.C(filter.Field).IsNotNull())
			continue
		}

		if len(filter.Values) == 0 {
			continue
		}
		switch filter.Operator {
		case criteria.Equal:
			expressions = append(expressions, goqu.C(filter.Field).Eq(filter.Values[0]))
		case criteria.NotEqual:
			expressions = append(expressions, goqu.C(filter.Field).Neq(filter.Values[0]))
		case criteria.LessThan:
			expressions = append(expressions, goqu.C(filter.Field).Lt(filter.Values[0]))
		case criteria.LessThanOrEqualTo:
			expressions = append(expressions, goqu.C(filter.Field).Lte(filter.Values[0]))
		case criteria.GreaterThan:
			expressions = append(expressions, goqu.C(filter.Field).Gt(filter.Values[0]))
		case criteria.GreaterThanOrEqualTo:
			expressions = append(expressions, goqu.C(filter.Field).Gte(filter.Values[0]))
		case criteria.In:
			expressions = append(expressions, goqu.C(filter.Field).In(filter.Values...))
		case criteria.NotIn:
			expressions = append(expressions, goqu.C(filter.Field).NotIn(filter.Values...))
		case criteria.Between:
			if len(filter.Values) != 2 {
				continue
			}
			expressions = append(expressions, goqu.C(filter.Field).Between(goqu.Range(filter.Values[0], filter.Values[1])))
		case criteria.NotBetween:
			if len(filter.Values) != 2 {
				continue
			}
			expressions = append(expressions, goqu.C(filter.Field).NotBetween(goqu.Range(filter.Values[0], filter.Values[1])))
		case criteria.Like:
			expressions = append(expressions, goqu.C(filter.Field).Like(filter.Values[0]))
		case criteria.ILike:
			expressions = append(expressions, goqu.C(filter.Field).ILike(filter.Values[0]))
		case criteria.NotLike:
			expressions = append(expressions, goqu.C(filter.Field).NotLike(filter.Values[0]))
		case criteria.NotILike:
			expressions = append(expressions, goqu.C(filter.Field).NotILike(filter.Values[0]))
		default:
		}
	}

	if len(expressions) == 0 {
		return nil, nil
	}

	var expressionList exp.ExpressionList
	if cr.Operator == criteria.Or {
		expressionList = goqu.Or(expressions...)
	} else {
		expressionList = goqu.And(expressions...)
	}
	return expressionList, nil
}
