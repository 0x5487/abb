package app

import (
	"github.com/jasonsoft/napnap"
)

func GetPaginationFromContext(c *napnap.Context) Pagination {
	page, err := c.QueryInt("page")
	if err != nil || page <= 0 {
		page = 1
	}
	perPage, err := c.QueryInt("per_page")
	if err != nil || perPage <= 0 || perPage > 100 {
		perPage = 25
	}

	pagination := Pagination{
		Page:    page,
		PerPage: perPage,
	}
	return pagination
}
