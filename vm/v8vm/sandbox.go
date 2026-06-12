package v8

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

const resultMaxLength = 65536 // UTF-16 code units, matching JS String.prototype.length

// jsStringLength returns the number of UTF-16 code units in s, which is what
// JavaScript reports for string.length.
func jsStringLength(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// Error message
var (
	ErrResultTooLong = errors.New("result too long")
)

// Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id       int // nolint
	flags    int64
	rt       *goja.Runtime
	host     *host.Host
	gasUsed  int64
	gasLimit int64
	deadline time.Time
}

// NewSandbox generate new sandbox for VM.
func NewSandbox(e *VM, flags int64) *Sandbox {
	rt := goja.New()

	s := &Sandbox{
		rt:    rt,
		flags: flags,
	}
	s.Init(e.vmType)
	return s
}

// Release release sandbox resources.
func (sbx *Sandbox) Release() {
	if sbx.rt != nil {
		sbx.rt.ClearInterrupt()
		sbx.rt = nil
	}
}

// Init add system functions.
func (sbx *Sandbox) Init(vmType vmPoolType) {
	rt := sbx.rt

	if vmType == CompileVMPool {
		// Compile VM only needs _cLog (stub) and the compile-time libraries.
		rt.Set("_cLog", func(goja.FunctionCall) goja.Value { return goja.Undefined() })
		rt.Set("console", map[string]func(goja.FunctionCall) goja.Value{
			"log":   func(goja.FunctionCall) goja.Value { return goja.Undefined() },
			"debug": func(goja.FunctionCall) goja.Value { return goja.Undefined() },
			"info":  func(goja.FunctionCall) goja.Value { return goja.Undefined() },
			"warn":  func(goja.FunctionCall) goja.Value { return goja.Undefined() },
			"error": func(goja.FunctionCall) goja.Value { return goja.Undefined() },
		})

		compileBootstrap := fmt.Sprintf(
			"let exports = {};\n"+
				"let module = {};\n"+
				"module.exports = {};\n"+
				"%s\n"+
				"const esprima = module.exports;\n"+
				"%s\n"+
				"const escodegen = module.exports;\n"+
				"%s\n"+
				"%s\n",
			esprimaJS, escodegenJS, validateJS, injectGasJS,
		)
		if _, err := rt.RunString(compileBootstrap); err != nil {
			panic(fmt.Sprintf("load compile bootstrap error: %v", err))
		}
		return
	}

	// Run VM: inject all host bindings.
	rt.Set("_cLog", newCLog(sbx))

	// Native constructors.
	rt.Set("IOSTBlockchain", func(call goja.ConstructorCall) *goja.Object {
		return newIOSTBlockchain(sbx)
	})
	rt.Set("IOSTStorage", func(call goja.ConstructorCall) *goja.Object {
		return newIOSTStorage(sbx)
	})
	rt.Set("IOSTInstruction", func(call goja.ConstructorCall) *goja.Object {
		return newIOSTInstruction(sbx)
	})
	rt.Set("_IOSTCrypto", func(call goja.ConstructorCall) *goja.Object {
		return newIOSTCrypto(sbx)
	})

	// Load runtime base libraries in one script.
	runtimeBootstrap := fmt.Sprintf(
		"let module = {};\n"+
			"module.exports = {};\n"+
			"%s\n"+
			"%s\n"+
			"let BigNumber = module.exports;\n"+
			"%s\n"+
			"%s\n"+
			"%s\n"+
			"%s\n",
		jsonJS, bignumberJS, int64JS, float64JS, utilsJS, consoleJS,
	)
	if _, err := rt.RunString(runtimeBootstrap); err != nil {
		panic(fmt.Sprintf("load runtime bootstrap error: %v", err))
	}

	// Pre-load storage and blockchain modules into global cache.
	for _, mod := range []struct{ name, src string }{{"storage", storageJS}, {"blockchain", blockchainJS}} {
		jsCode := fmt.Sprintf(`
(function(){
var module = { exports: {} };
(function(exports, require, module, __filename, __dirname) {
%s
})(module.exports, null, module, '%s.js', '');
globalThis.__modules_%s = module.exports;
})();
`, mod.src, mod.name, mod.name)
		if _, err := rt.RunString(jsCode); err != nil {
			panic(fmt.Sprintf("load %s module error: %v", mod.name, err))
		}
	}

	// Inject pure-JS require/run helpers that use the pre-loaded modules.
	if _, err := rt.RunString(`
_native_require = function(id) { return ''; };
_native_run = function(source, filename) {
	return function(exports, require, module, __filename, __dirname) {
		if (filename === 'storage.js') {
			module.exports = globalThis.__modules_storage;
		} else if (filename === 'blockchain.js') {
			module.exports = globalThis.__modules_blockchain;
		}
	};
};
`); err != nil {
		panic(fmt.Sprintf("inject require helpers error: %v", err))
	}

	// Load vm.js.
	if _, err := rt.RunString(vmJS); err != nil {
		panic(fmt.Sprintf("load vm.js error: %v", err))
	}

	// Remove _native_run from global scope after vm.js has captured it.
	if _, err := rt.RunString("_native_run = undefined;"); err != nil {
		panic(fmt.Sprintf("remove _native_run error: %v", err))
	}

	// Polyfill missing globals before environment.js tries to null them.
	if _, err := rt.RunString(`
var _native_log = function(){};
if (typeof Atomics === 'undefined') { var Atomics = {}; }
if (typeof Intl === 'undefined') { var Intl = {}; }
if (typeof WebAssembly === 'undefined') { var WebAssembly = {}; }
if (typeof SharedArrayBuffer === 'undefined') { var SharedArrayBuffer = {}; }
if (typeof DataView === 'undefined') { var DataView = {}; }
`); err != nil {
		panic(fmt.Sprintf("inject missing globals error: %v", err))
	}

	// Capture methods that environment.js sets to null so we can restore them.
	if _, err := rt.RunString(`
var __savedArrayFrom = Array.from;
var __savedArrayOf = Array.of;
var __savedFunction = Function;
`); err != nil {
		panic(fmt.Sprintf("capture globals error: %v", err))
	}

	// Load environment.js.
	if _, err := rt.RunString(environmentJS); err != nil {
		panic(fmt.Sprintf("load environment.js error: %v", err))
	}

	// Restore Array.from and Array.of.
	if _, err := rt.RunString(`
Array.from = __savedArrayFrom;
Array.of = __savedArrayOf;
`); err != nil {
		panic(fmt.Sprintf("restore Array methods error: %v", err))
	}

	// Block dangerous dynamic-code features.
	if _, err := rt.RunString(`
var __blockedEval = function() { throw new Error("Code generation from strings disallowed for this context"); };
globalThis.eval = __blockedEval;

var __origFunction = __savedFunction;
var __origProto = __origFunction.prototype;
var __blockedFunction = function() { throw new Error("Function is not a constructor"); };
__origProto.constructor = __blockedFunction;
globalThis.Function = __blockedFunction;
`); err != nil {
		panic(fmt.Sprintf("block eval/function error: %v", err))
	}

	// Re-inject the native constructors that environment.js has nulled out.
	if _, err := rt.RunString(`
IOSTBlockchain = function() { return new IOSTBlockchain(); };
IOSTStorage    = function() { return new IOSTStorage(); };
IOSTInstruction= function() { return new IOSTInstruction(); };
_IOSTCrypto    = function() { return new _IOSTCrypto(); };
`); err != nil {
		panic(fmt.Sprintf("re-inject native constructors error: %v", err))
	}

	// Wrap Array constructor to reject pathologically large allocations,
	// matching the old V8 memory-limit behavior.
	if _, err := rt.RunString(`
const __origArray = Array;
const __wrappedArray = function() {
    if (arguments.length === 1 && typeof arguments[0] === 'number' && arguments[0] > 100000000) {
        throw new Error("out of memory");
    }
    if (arguments.length === 1 && typeof arguments[0] === 'number' && arguments[0] < 0) {
        throw new RangeError("Invalid array length");
    }
    return __origArray.apply(this, arguments);
};
__wrappedArray.prototype = __origArray.prototype;
Object.setPrototypeOf(__wrappedArray, __origArray);
Array = __wrappedArray;
`); err != nil {
		panic(fmt.Sprintf("wrap Array constructor error: %v", err))
	}
}

// GetFlags ...
func (sbx *Sandbox) GetFlags() int64 {
	return sbx.flags
}

// SetGasLimit set gas limit in context.
func (sbx *Sandbox) SetGasLimit(limit int64) {
	sbx.gasUsed = 0
	sbx.gasLimit = limit
}

// SetHost set host in sandbox and set gas limit.
func (sbx *Sandbox) SetHost(host *host.Host) {
	sbx.host = host
	sbx.SetGasLimit(host.GasLimitValue())
}

// Validate contract before save, return err if invalid.
func (sbx *Sandbox) Validate(contract *contract.Contract) error {
	code := moduleReplacer.Replace(contract.Code)
	abi, _ := json.Marshal(contract.Info.Abi)

	jsCode := "(function(){\nconst source = \"" + code + "\";\nconst abi = " + string(abi) + ";\nreturn validate(source, abi);\n})();"
	result, err := sbx.rt.RunString(jsCode)
	resStr := ""
	if err == nil {
		resStr = result.String()
	}
	if err != nil || resStr != "success" {
		if err == nil {
			err = errors.New("")
		}
		return fmt.Errorf("validate code error: %v, result: %v", err, resStr)
	}
	return nil
}

// Compile contract before execution, return compiled code.
func (sbx *Sandbox) Compile(contract *contract.Contract) (string, error) {
	code := moduleReplacer.Replace(contract.Code)

	jsCode := "(function(){\nconst source = \"" + code + "\";\nreturn injectGas(source);\n})();"
	result, err := sbx.rt.RunString(jsCode)
	if err != nil {
		return "", err
	}
	rs := result.String()
	if rs == "" {
		return "", errors.New("inject gas failed")
	}
	return rs, nil
}

// Prepare for contract, inject code.
func (sbx *Sandbox) Prepare(contract *contract.Contract, function string, args []any) (string, error) {
	code := contract.Code

	if function == "constructor" {
		return fmt.Sprintf(`
%s
var _obj = new module.exports;

var ret = 0;
return ret;
`, code), nil
	}

	argStr, err := formatFuncArgs(args)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
%s;
const _obj = new module.exports;

// run contract with specified function and args
let rs = _obj.%s(%s);
if ((typeof rs === 'function') || (typeof rs === 'object')) {
	_IOSTInstruction_counter.incr(12);
	try {
		rs = JSON.stringify(rs);
	} catch (e) {
		throw new Error('JSON.stringify failed: ' + (e && e.message ? e.message : e));
	}
}
if (typeof rs === 'string' && rs.length > %d) {
	throw new Error("result too long");
}
return rs;
`, code, function, argStr, resultMaxLength), nil
}

// Execute prepared code, return results, gasUsed.
func (sbx *Sandbox) Execute(preparedCode string) (string, int64, error) {
	now := time.Now()
	if !sbx.host.Deadline().After(now) {
		return "", 0, errors.New("execution killed")
	}

	// Reset gas used so that Init-time library loading gas is not counted.
	sbx.gasUsed = 0
	sbx.gasLimit = sbx.host.GasLimitValue()
	sbx.deadline = sbx.host.Deadline()

	// Clear any stale interrupt flag before this execution.
	sbx.rt.ClearInterrupt()

	// Start a watcher that interrupts if the deadline is exceeded.
	d := time.Until(sbx.deadline)
	var timer *time.Timer
	if d > 0 {
		timer = time.AfterFunc(d, func() {
			if sbx.rt != nil {
				sbx.rt.Interrupt("execution killed")
			}
		})
		defer timer.Stop()
	} else {
		return "", 0, errors.New("execution killed")
	}

	// Preload block/tx globals.
	if _, err := sbx.rt.RunString(`
const blockInfo = JSON.parse(blockchain.blockInfo());
const block = {
   number: blockInfo.number,
   parentHash: blockInfo.parent_hash,
   witness: blockInfo.witness,
   time: blockInfo.time
};

const txInfo = JSON.parse(blockchain.txInfo());
const tx = {
   time: txInfo.time,
   hash: txInfo.hash,
   expiration: txInfo.expiration,
   gasLimit: txInfo.gas_limit,
   gasRatio: txInfo.gas_ratio,
   authList: txInfo.auth_list,
   publisher: txInfo.publisher
};
`); err != nil {
		return "", sbx.gasUsed, err
	}

	sbx.gasUsed = 0

	// Wrap in IIFE with try-catch so thrown primitives are converted to Error objects.
	wrappedCode := fmt.Sprintf(`
(function() {
    try {
        %s
    } catch (e) {
        if (e instanceof Error) {
            throw e;
        } else if (typeof e === 'string') {
            throw new Error(e);
        } else {
            throw new Error(String(e));
        }
    }
})();
`, preparedCode)

	result, err := sbx.rt.RunString(wrappedCode)

	if err != nil {
		return "", sbx.gasUsed, err
	}

	var str string
	if goja.IsUndefined(result) {
		str = ""
	} else if goja.IsNull(result) {
		str = "null"
	} else if s, ok := result.Export().(string); ok {
		str = s
	} else {
		str = result.String()
	}

	if jsStringLength(str) > resultMaxLength {
		return "", sbx.gasUsed, ErrResultTooLong
	}

	return str, sbx.gasUsed, nil
}

func formatFuncArgs(args []any) (string, error) {
	var strArgs []string
	for _, arg := range args {
		switch v := arg.(type) {
		case []byte:
			strArgs = append(strArgs, string(v))
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			strArgs = append(strArgs, string(b))
		}
	}
	argStr := strings.Join(strArgs, ",")

	return argStr, nil
}
