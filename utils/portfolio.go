package utils


func IfThenElse[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func IfThenElseSlice[T any](cond bool, a, b []T) []T {
	if cond {
		return a
	}
	return b
}