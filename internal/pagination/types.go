// Package pagination provides reusable types and utilities for paginated
// list queries. Every list endpoint in the project should use this package
// instead of defining its own page/page-size/total/total-pages logic.
package pagination

// Paging carries the page and page-size parameters for a list query.
// Validation / normalisation is done by Normalise or NewPaged.
type Paging struct {
	Page     int
	PageSize int
}

// Paged wraps a paginated result set with the metadata needed by the
// frontend to render a paginator.
type Paged[T any] struct {
	Items      []T  `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// NewPaged builds a Paged[T] from a result slice, the total count, and the
// original Paging.  Page and PageSize are assumed to be already normalised
// (call Normalise before this if the values come from untrusted input).
func NewPaged[T any](items []T, total int64, paging Paging) *Paged[T] {
	totalPages := totalPages(total, paging.PageSize)
	return &Paged[T]{
		Items:      items,
		Total:      total,
		Page:       paging.Page,
		PageSize:   paging.PageSize,
		TotalPages: totalPages,
	}
}

func totalPages(total int64, pageSize int) int {
	if pageSize < 1 || total < 1 {
		return 0
	}
	pages := int(total / int64(pageSize))
	if total%int64(pageSize) != 0 {
		pages++
	}
	return pages
}
