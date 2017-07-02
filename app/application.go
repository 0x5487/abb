package app

import "fmt"

type AppError struct {
	ErrorCode string `json:"error_code" bson:"-"`
	Message   string `json:"message" bson:"message"`
}

func (e AppError) Error() string {
	return fmt.Sprintf("%s - %s", e.ErrorCode, e.Message)
}

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPage  int `json:"total_page"`
}

func (p *Pagination) SetTotalCount(total int) {
	p.TotalCount = total
	p.TotalPage = p.TotalCount / p.PerPage
	if p.TotalCount > 0 && p.TotalPage == 0 {
		p.TotalPage = 1
	} else {
		mod := p.TotalCount % p.PerPage
		if mod > 0 {
			p.TotalPage++
		}
	}
}

func (p *Pagination) Skip() int {
	return (p.Page - 1) * p.PerPage
}

type ApiPagiationResult struct {
	Pagination Pagination  `json:"meta"`
	Data       interface{} `json:"data"`
}

type ApiCollectionResult struct {
	TotalCount int         `json:"total_count"`
	Result     interface{} `json:"result"`
}
