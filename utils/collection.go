package utils

func RemoveDuplicate[T string | int](tSlice []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range tSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
