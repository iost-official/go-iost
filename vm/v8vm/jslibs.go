package v8

import (
	_ "embed"
)

//go:embed libjs/json.js
var jsonJS string

//go:embed libjs/bignumber.js
var bignumberJS string

//go:embed libjs/int64.js
var int64JS string

//go:embed libjs/float64.js
var float64JS string

//go:embed libjs/utils.js
var utilsJS string

//go:embed libjs/console.js
var consoleJS string

//go:embed libjs/vm.js
var vmJS string

//go:embed libjs/environment.js
var environmentJS string

//go:embed libjs/storage.js
var storageJS string

//go:embed libjs/blockchain.js
var blockchainJS string

//go:embed libjs/esprima.js
var esprimaJS string

//go:embed libjs/escodegen.js
var escodegenJS string

//go:embed libjs/validate.js
var validateJS string

//go:embed libjs/inject_gas.js
var injectGasJS string

// runtimeLibs are loaded into every run VM.
var runtimeLibs = []struct {
	name string
	src  string
}{
	{"json.js", jsonJS},
	{"bignumber.js", bignumberJS},
	{"int64.js", int64JS},
	{"float64.js", float64JS},
	{"utils.js", utilsJS},
	{"console.js", consoleJS},
}

// compileLibs are concatenated and loaded into every compile VM.
var compileLibs = []struct {
	name string
	src  string
}{
	{"esprima.js", esprimaJS},
	{"escodegen.js", escodegenJS},
	{"validate.js", validateJS},
	{"inject_gas.js", injectGasJS},
}
