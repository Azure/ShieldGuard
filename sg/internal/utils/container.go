package utils

// Map maps a list with a function.
func Map[T any, U any](xs []T, fn func(T) U) []U {
	ys := make([]U, len(xs))
	for i, item := range xs {
		ys[i] = fn(item)
	}
	return ys
}

// Filter filters a list with a predicate.
func Filter[T any](xs []T, pred func(T) bool) []T {
	ys := make([]T, 0)
	for _, x := range xs {
		if pred(x) {
			ys = append(ys, x)
		}
	}
	return ys
}
