package common

func PopFront[T any](elems *[]T) (T, bool) {
	if len(*elems) == 0 {
		var zero T
		return zero, false
	}
	e := (*elems)[0]
	*elems = (*elems)[1:]
	return e, true
}
