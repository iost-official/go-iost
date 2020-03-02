// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
)

var template = `package contract

import "net/url"

const contractJSEscaped = "%s"

// CompiledContract is the compiled js code
var CompiledContract string

func init() {
	CompiledContract, _ = url.QueryUnescape(contractJSEscaped)
}
`

func main() {
	contractJS, err := ioutil.ReadFile("./dist/bundle.js")
	if err != nil {
		log.Fatal(err)
	}
	contractJSEscaped := url.QueryEscape(string(contractJS))
	content := fmt.Sprintf(template, contractJSEscaped)
	err = ioutil.WriteFile("./compiledContract.go", []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
