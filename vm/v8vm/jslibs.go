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
