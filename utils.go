package queue

func Remove[T any](slice []T, removeIf func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if !removeIf(v) {
			result = append(result, v)
		}
	}
	return result
}
