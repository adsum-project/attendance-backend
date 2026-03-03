package pagination

import "context"

const (
	DefaultPerPage = 10
	MaxPerPage     = 100
)

// Result holds a single page of data plus pagination metadata.
type Result[T any] struct {
	Data    []T // items for the current page
	Total   int // total count across all pages
	Page    int // page number used (1-based)
	PerPage int // page size used
}

func Paginate[T any](ctx context.Context, page, perPage int, fetch func(context.Context, int, int) ([]T, error), count func(context.Context) (int, error)) (*Result[T], error) {
	page, perPage = Normalize(page, perPage)
	data, err := fetch(ctx, page, perPage)
	if err != nil {
		return nil, err
	}
	total, err := count(ctx)
	if err != nil {
		return nil, err
	}
	return &Result[T]{Data: data, Total: total, Page: page, PerPage: perPage}, nil
}

// BindID adapts repo methods that take an id (e.g. GetClasses(ctx, moduleID, page, perPage))
// to the fetch/count signatures expected by Paginate. Use when the repo needs an extra id parameter.
func BindID[T any](id string, fetch func(context.Context, string, int, int) ([]T, error), count func(context.Context, string) (int, error)) (func(context.Context, int, int) ([]T, error), func(context.Context) (int, error)) {
	return func(ctx context.Context, p, pp int) ([]T, error) { return fetch(ctx, id, p, pp) },
		func(ctx context.Context) (int, error) { return count(ctx, id) }
}

// BindIDAndMap is like BindID but applies mapFn to fetch results. Use when the repo returns a different type
// than the service (e.g. repo returns []CourseStudent, service returns []CourseStudentEnrollment after enrichment).
func BindIDAndMap[T, U any](id string, fetch func(context.Context, string, int, int) ([]T, error), count func(context.Context, string) (int, error), mapFn func(context.Context, []T) ([]U, error)) (func(context.Context, int, int) ([]U, error), func(context.Context) (int, error)) {
	return func(ctx context.Context, p, pp int) ([]U, error) {
			data, err := fetch(ctx, id, p, pp)
			if err != nil {
				return nil, err
			}
			return mapFn(ctx, data)
		},
		func(ctx context.Context) (int, error) { return count(ctx, id) }
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
