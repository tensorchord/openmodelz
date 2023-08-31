package util

func Ptr[T any](v T) *T {
	return &v
}

func PtrCopy[T any](v T) *T {
	n := new(T)
	*n = v
	return n
}
