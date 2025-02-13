package sql

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/criteria"
	"github.com/hadroncorp/geck/structs"
)

// ExecCriteriaParams are the parameter for [ExecCriteria].
type ExecCriteriaParams struct {
	DB              *goqu.Database
	Table           string
	Criteria        criteria.Criteria
	FieldTranslator *persistence.FieldTranslator
}

// ExecCriteria executes a query based on the given [ExecCriteriaParams.Criteria] object.
func ExecCriteria[T any](ctx context.Context, params ExecCriteriaParams) ([]T, error) {
	if params.Criteria.PageToken != nil {
		// page token replaces some criteria fields
		params.Criteria.Sorting.Field = params.Criteria.PageToken.Sort.Field
		params.Criteria.Sorting.Operator = lo.CoalesceOrEmpty(
			criteria.NewSortOperator(params.Criteria.PageToken.Sort.Operator))

		if params.Criteria.HasNextPageToken() {
			params.Criteria.Filters = append(params.Criteria.Filters, criteria.Filter{
				Field:    params.Criteria.PageToken.CursorName,
				Operator: criteria.GreaterThan,
				Values:   []any{params.Criteria.PageToken.EndCursor},
			})
		} else if params.Criteria.HasPreviousPageToken() {
			params.Criteria.Filters = append(params.Criteria.Filters, criteria.Filter{
				Field:    params.Criteria.PageToken.CursorName,
				Operator: criteria.LessThan,
				Values:   []any{params.Criteria.PageToken.StartCursor},
			})
		}
	}

	if err := criteria.TranslateFields(params.FieldTranslator, &params.Criteria); err != nil {
		return nil, err
	}

	expressionList, err := newFilterQuery(params.Criteria)
	if err != nil {
		return nil, err
	}

	var orderList exp.OrderedExpression
	if params.Criteria.HasSorting() && params.Criteria.Sorting.Operator == criteria.SortAscending {
		orderList = goqu.C(params.Criteria.Sorting.Field).Asc()
	} else if params.Criteria.HasSorting() && params.Criteria.Sorting.Operator == criteria.SortDescending {
		orderList = goqu.C(params.Criteria.Sorting.Field).Desc()
	}

	dataset := params.DB.From(params.Table).
		Limit(uint(params.Criteria.PageSize)).
		Order(orderList)
	if expressionList != nil && !expressionList.IsEmpty() {
		dataset = dataset.Where(expressionList)
	}

	var models []T
	err = dataset.ScanStructsContext(ctx, &models)
	if err != nil {
		return nil, err
	}

	if len(models) > 0 && params.Criteria.HasPreviousPageToken() && !params.Criteria.HasInitialSort(criteria.SortDescending) {
		slices.SortFunc(models, func(a, b T) int {
			sortValA := structs.GetStructValue(a, params.Criteria.Sorting.Field,
				structs.WithTag("db"))
			sortValB := structs.GetStructValue(b, params.Criteria.Sorting.Field,
				structs.WithTag("db"))
			return strings.Compare(fmt.Sprintf("%v", sortValA), fmt.Sprintf("%v", sortValB))
		})
	} else if len(models) > 0 && params.Criteria.HasNextPageToken() && params.Criteria.HasInitialSort(criteria.SortDescending) {
		slices.SortFunc(models, func(a, b T) int {
			sortValA := structs.GetStructValue(a, params.Criteria.Sorting.Field,
				structs.WithTag("db"))
			sortValB := structs.GetStructValue(b, params.Criteria.Sorting.Field,
				structs.WithTag("db"))
			return strings.Compare(fmt.Sprintf("%v", sortValB), fmt.Sprintf("%v", sortValA))
		})
	}
	return models, nil
}
