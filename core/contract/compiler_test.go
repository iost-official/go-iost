package contract

import (
	"fmt"
	"testing"
)

var raw = `{
"lang": "javascript",
"version": "1.0.0",
"abi": [{
	"name": "abc",
	"args": ["", "", ""]
	}, {
		"name": "def",
	"args": ["string", "string", "number"],
	"payment": 2,
	"limit": {
		"data": 1,
		"net": 2,
		"cpu": 3
	},
	"gas_price": 123
	}]
}
`

func TestCompiler_ParseInfo(t *testing.T) {
	var compiler Compiler
	info, err := compiler.parseInfo(raw)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("info: %+v\n", info)

	if info.Lang != "javascript" || info.Version != "1.0.0" {
		t.Fatal(info)
	}
}
