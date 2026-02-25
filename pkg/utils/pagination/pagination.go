package pagination

import "context"

const (
	DefaultPerPage = 10
	MaxPerPage     = 100
)

func Paginate[T any](ctx context.Context, page, perPage int, fetch func(context.Context, int, int) ([]T, error), count func(context.Context) (int, error)) (data []T, total int, usedPage, usedPerPage int, err error) {
	page, perPage = Normalize(page, perPage)
	data, err = fetch(ctx, page, perPage)
	if err != nil {
		return
	}
	total, err = count(ctx)
	if err != nil {
		return
	}
	usedPage, usedPerPage = page, perPage
	return
}

func Normalize(page, perPage int) (int, int) {
	return NormalizeWithLimits(page, perPage, DefaultPerPage, MaxPerPage)
}

func NormalizeWithLimits(page, perPage, defaultPerPage, maxPerPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = defaultPerPage
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage
}
