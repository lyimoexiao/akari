package pagination

// DefaultPageSize is the default number of items per page when none is
// specified.
const DefaultPageSize = 20

// MaxPageSize is the upper bound for page size values.
const MaxPageSize = 100

// Normalise clamps Page to [1, ∞) and PageSize to [1, MaxPageSize].
// When p.PageSize is 0 it is set to DefaultPageSize.
// The original Paging is returned (not a copy) — callers that need the
// original must copy it first.
func (p *Paging) Normalise() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	} else if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
}
