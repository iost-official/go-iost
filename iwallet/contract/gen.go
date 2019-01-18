// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

func main() {
	fi, err := os.OpenFile("./compiledContract.go", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	contractJS, err := ioutil.ReadFile("./dist/bundle.js")
	if err != nil {
		log.Fatal(err)
	}

	contractJSEscaped := url.QueryEscape(string(contractJS))

	fi.WriteString(fmt.Sprintf(`package contract

import "net/url"

const contractJSEscaped = "%s"

var CompiledContract string

func init() {
	CompiledContract, _ = url.QueryUnescape(contractJSEscaped)
}`, contractJSEscaped))
}
