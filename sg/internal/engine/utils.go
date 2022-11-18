package engine

func filterList[T any](xs []T, pred func(T) bool) []T {
	var ys []T
	for _, x := range xs {
		if pred(x) {
			ys = append(ys, x)
		}
	}
	return ys
}
