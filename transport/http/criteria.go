package http

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence/criteria"
	"github.com/hadroncorp/geck/persistence/paging/pagetoken"
)

// NewCriteriaQuery allocates a [criteria.Query] based on an HTTP request (i.e. `c`, an [echo.Context]).
func NewCriteriaQuery(c echo.Context) (criteria.Query, error) {
	var pageSize int64
	if rawSize := c.QueryParam("page_size"); rawSize != "" {
		pageSize, _ = strconv.ParseInt(rawSize, 10, 64)
	}

	pageToken, err := pagetoken.UnmarshalEmptyable(c.QueryParam("page_token"))
	if err != nil {
		return criteria.Query{}, err
	}

	var sortQuery criteria.SortQuery
	sortField := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")
	if sortField != "" && sortOrder != "" {
		sortQuery.Field = sortField
		sortQuery.Operator = sortOrder
	}
	return criteria.Query{
		PageSize:  lo.If(pageSize > 0 && pageSize <= 250, pageSize).Else(25),
		PageToken: pageToken,
		Sort:      sortQuery,
	}, nil
}
