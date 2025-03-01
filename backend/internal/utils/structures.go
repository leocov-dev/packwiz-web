package utils

func DeepCopyMapStringSlice(src map[string][]string) map[string][]string {
	dst := make(map[string][]string)

	for key, slice := range src {
		copiedSlice := make([]string, len(slice))
		copy(copiedSlice, slice)

		dst[key] = copiedSlice
	}

	return dst
}
