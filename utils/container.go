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

func FilterItem(A, B []string) []string {
	existingEl := make(map[string]bool)

	for _, el := range B {
		existingEl[el] = true
	}

	var nonExistingEl []string

	for _, el := range A {
		if !existingEl[el] {
			nonExistingEl = append(nonExistingEl, el)
		}
	}
	return nonExistingEl
}


