package rpc

import (
	"fmt"
	"strings"
)

func checkHashValid(hash string) error {
	maxLen := 100
	if len(hash) >= maxLen {
		return fmt.Errorf("hash len(%v) is greater then %v", len(hash), maxLen)
	}
	return nil
}

func isEthAddress(id string) bool {
	if strings.HasPrefix(id, "0x") && len(id) == 42 && strings.ToLower(id) == id {
		return true
	}
	return false
}

func checkIDValid(id string) error {
	if isEthAddress(id) {
		return nil
	}
	if strings.HasPrefix(id, "Contract") || strings.HasSuffix(id, ".iost") {
		if len(id) >= 100 {
			return fmt.Errorf("id invalid. ContractID length should be less then 100 - %v", len(id))
		}
		return nil
	}

	if len(id) < 5 || len(id) > 11 {
		return fmt.Errorf("id invalid. id length should be between 5,11 - %v ", len(id))
	}

	for _, v := range id {
		if !((v >= 'a' && v <= 'z') || (v >= '0' && v <= '9' || v == '_')) {
			return fmt.Errorf("id invalid. id contains invalid character - %v", v)
		}
	}
	return nil
}
