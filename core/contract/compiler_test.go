package contract

import (
	"testing"

	"github.com/bitly/go-simplejson"
)

var raw = `{
"lang": "javascript",
"version": "1.0.0",
"abi": [{
	"name": "abc",
	"args": ["", "", ""],
	"payment": 0,
	"cost_limit": [1,2,3], 
	"price_limit": 123
	}, {
		"name": "def",
	"args": ["string", "string", "number"],
	"payment": 2,
	"cost_limit": [1,2,3],
	"price_limit": 123
	}]
}
`

func TestCompiler_ParseABI(t *testing.T) {

	json, err := simplejson.NewJson([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	var compiler Compiler
	abis, err := compiler.parseABI(json.Get("abi"))
	if err != nil {
		t.Fatal(err)
	}
	if len(abis) != 2 {
		t.Fatal(abis)
	}
}

func TestCompiler_ParseInfo(t *testing.T) {
	json, err := simplejson.NewJson([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	var compiler Compiler
	info, err := compiler.parseInfo(json)

	if info.Lang != "javascript" || info.VersionCode != "1.0.0" {
		t.Fatal(info)
	}
}
