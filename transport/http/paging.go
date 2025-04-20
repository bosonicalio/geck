package http

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence/paging"
)

// NewPaginationOptions creates a new set of pagination options based on the query parameters
// from the given echo context. It uses the "page_size" and "page_token" query parameters
// to set the limit and page token for pagination.
func NewPaginationOptions(c echo.Context) []paging.Option {
	limit, _ := strconv.Atoi(c.QueryParam("page_size"))
	return []paging.Option{
		paging.WithLimit(lo.CoalesceOrEmpty(limit, 25)),
		paging.WithPageToken(c.QueryParam("page_token")),
	}
}
