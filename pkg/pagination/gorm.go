package pagination

import "gorm.io/gorm"

// ApplyOffsetLimit returns a GORM scopes-compatible function that applies
// skip/limit based on a (normalised) Paging value.
//
// Usage:
//
//	var p paging.Paging
//	p.Normalise()
//	dbQuery.Scopes(pagination.ApplyOffsetLimit(p)).Find(&items)
func ApplyOffsetLimit(p Paging) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((p.Page - 1) * p.PageSize).Limit(p.PageSize)
	}
}
