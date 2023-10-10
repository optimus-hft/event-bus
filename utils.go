package event_bus

// RemoveFromSlice removes an element with index i from the slice.
func RemoveFromSlice[S interface{ ~[]E }, E any](s S, i int) S {
	result := make([]E, len(s)-1)
	copy(result[:i], s[:i])
	copy(result[i:], s[i+1:])

	return result
}

func FindSliceIndex[T comparable](haystack []T, needle T) int {
	index := -1

	for i, item := range haystack {
		if item == needle {
			index = i

			break
		}
	}

	return index
}
