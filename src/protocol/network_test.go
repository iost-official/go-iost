package protocol

import (
	"IOS/src/iosbase"
	"fmt"
	"testing"
)

type MockBlock struct {
	isDecodeSuccess bool
}

func TestRecorderFilter(t *testing.T) {

	var req chan iosbase.Request
	var res chan iosbase.Response

	fmt.Println("")

}
