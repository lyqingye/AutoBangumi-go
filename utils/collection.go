package utils

func RemoveDuplicate[T string | int](tSlice []T) []T {
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range tSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

// Difference c1 - c2
func Difference[T string | int](c1 []T, c2 []T) []T {
	allKeys := make(map[T]bool)
	for _, item := range c1 {
		allKeys[item] = true
	}

	var ret []T
	for _, item := range c2 {
		if _, found := allKeys[item]; !found {
			ret = append(ret, item)
		}
	}
	return ret
}

func Keys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for k, _ := range m {
		keys = append(keys, k)
	}
	return keys
}
