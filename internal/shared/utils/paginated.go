package utils

import (
	"math"
	"strings"
)

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

type Pagination struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
	Search    string `json:"search"`
}

func NewPagination(page, pageSize int, sort, direction, search string) Pagination {
	if page < 1 {
		page = defaultPage
	}

	if pageSize < 1 {
		pageSize = defaultPageSize
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	if sort == "" {
		sort = "created_at"
	}

	direction = strings.ToUpper(direction)
	if direction != "ASC" && direction != "DESC" {
		direction = "DESC"
	}

	search = strings.TrimSpace(search)

	return Pagination{
		Page:      page,
		PageSize:  pageSize,
		Sort:      sort,
		Direction: direction,
		Search:    search,
	}
}

func (p Pagination) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Pagination) GetLimit() int {
	return p.PageSize
}

func (p Pagination) GetSort() string {
	return p.Sort
}

func (p Pagination) GetDirection() string {
	return p.Direction
}

type PaginatedResult[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginatedResult[T any](items []T, total int64, p Pagination) *PaginatedResult[T] {
	if items == nil {
		items = []T{}
	}

	totalPages := int(math.Ceil(float64(total) / float64(p.PageSize)))

	return &PaginatedResult[T]{
		Items:      items,
		TotalCount: total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}
}
