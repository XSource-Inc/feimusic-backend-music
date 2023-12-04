package utils

import "context"

func AddToMapIfNotNil[K comparable, V any, M ~map[K]any](m M, v *V, k K) {
	if v == nil {
		return
	}
	m[k] = *v
} 

func GetValue(ctx context.Context, value string) string {
	if ctx.Value(value) != nil {
		val, ok := ctx.Value(value).(string)
		if ok {
			return val
		}
	}
	return ""
}

//TODO：待升级
func RemoveString(slice []string, target string) []string {
	result := slice[:0]

	for _, s := range slice {
		if s != target {
			result = append(result, s)
		}
	}

	return result
}

// A中存在，B中不存在的
func FilterItem[T comparable](A, B []T) []T {
	existingEl := make(map[T]bool)

	for _, el := range B {
		existingEl[el] = true
	}

	var nonExistingEl []T

	for _, el := range A {
		if !existingEl[el] {
			nonExistingEl = append(nonExistingEl, el)
		}
	}
	return nonExistingEl
}

// A、B的交集
func Intersection(A, B []int64) []int64{
	AEl := make(map[int64]bool)
	BEl := make(map[int64]bool)

	for _, el := range A{
		AEl[el] = true
	}
	for _, el := range B{
		BEl[el] = true
	}

	innerEl := []int64{}

	for el := range AEl{
		if ok := BEl[el]; ok{
			innerEl = append(innerEl, el)
		}
	} 

	return innerEl
}

