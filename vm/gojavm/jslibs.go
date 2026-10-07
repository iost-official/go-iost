package gojavm

import (
	_ "embed"

	"github.com/iost-official/go-iost/v3/vm/jslib"
)

// The consensus-critical libraries are shared with the V8 engine and live
// in vm/jslib. Only goja-specific runtime files are embedded from libjs.
var (
	jsonJS        = jslib.JSON
	bignumberJS   = jslib.BigNumber
	int64JS       = jslib.Int64
	float64JS     = jslib.Float64
	utilsJS       = jslib.Utils
	environmentJS = jslib.Environment
	blockchainJS  = jslib.Blockchain
	esprimaJS     = jslib.Esprima
	escodegenJS   = jslib.Escodegen
	validateJS    = jslib.Validate
	injectGasJS   = jslib.InjectGas
)

//go:embed libjs/console.js
var consoleJS string

//go:embed libjs/vm.js
var vmJS string

//go:embed libjs/storage.js
var storageJS string

//go:embed libjs/v8sort.js
var v8sortJS string
