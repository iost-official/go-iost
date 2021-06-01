package common

// StringSliceEqual is compare two strings
func StringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if (a == nil) != (b == nil) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// BelongsTo will judage if a public key is a witness.
func BelongsTo(elem string, arr []string) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

func AppendIfNotExists(arr *[]string, elem string) {
	found := false
	for _, item := range *arr {
		if item == elem {
			found = true
			break
		}
	}
	if !found {
		*arr = append(*arr, elem)
	}
}
