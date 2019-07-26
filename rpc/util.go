package rpc

import "fmt"

func checkHashValid(hash string) error {
	maxLen := 100
	if len(hash) >= maxLen {
		return fmt.Errorf("Hash len=[%v] is greater then %v", len(hash), maxLen)
	}
	return nil
}
