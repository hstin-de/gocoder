package utils

func MapToSlice[K comparable, V any](data map[K]V) []V {
	values := make([]V, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}
