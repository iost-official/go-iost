// Package jslib embeds the consensus-critical JavaScript runtime library
// sources shared by the V8 (vm/v8vm) and goja (vm/gojavm) engines.
//
// These files define the exact semantics every engine must execute
// identically, so they live in a single place. The V8 C++ build packs them
// into libvm via vm/v8vm/v8/Makefile (js_bin target); the goja engine
// embeds them from this package. Engine-specific runtime files are kept by
// each engine: vm/v8vm/v8/libjs (console.js, storage.js, vm.js) and
// vm/gojavm/libjs (console.js, storage.js, vm.js, v8sort.js).
package jslib

import _ "embed"

//go:embed json.js
var JSON string

//go:embed bignumber.js
var BigNumber string

//go:embed int64.js
var Int64 string

//go:embed float64.js
var Float64 string

//go:embed utils.js
var Utils string

//go:embed environment.js
var Environment string

//go:embed blockchain.js
var Blockchain string

//go:embed esprima.js
var Esprima string

//go:embed escodegen.js
var Escodegen string

//go:embed validate.js
var Validate string

//go:embed inject_gas.js
var InjectGas string
