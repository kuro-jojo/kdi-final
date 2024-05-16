package utils

func ModelArrayToStringArray[T any](items []T, getter func(T) string) []string {
	var result []string
	for _, item := range items {
		result = append(result, getter(item))
	}
	return result
}
