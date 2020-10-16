package blockcache

func SetInsert(arr *[]string, elem string) {
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
