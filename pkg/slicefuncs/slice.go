package slicefuncs

func FindIndexFunc[T any](o []T, compareFunc func(elem T) bool) int {
	for i, v := range o {
		if compareFunc(v) {
			return i
		}
	}
	return -1
}

func FilterFunc[T any](o []T, compareFunc func(elem T) bool) []T {
	newArray := []T{}

	for _, v := range o {
		if compareFunc(v) {
			newArray = append(newArray, v)
		}
	}

	return newArray
}

func DeleteIndex[T any](o []T, idx int) []T {
	newArray := o
	newArray[idx] = newArray[len(newArray)-1]
	return newArray[:len(newArray)-1]
}
