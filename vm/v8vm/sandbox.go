package v8

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Gaurav-Gosain/quickjs"
	"github.com/iost-official/go-iost/v3/core/contract"
	"github.com/iost-official/go-iost/v3/vm/host"
)

const resultMaxLength = 65536 // byte

// Error message
var (
	ErrResultTooLong = errors.New("result too long")
)

// Sandbox is an execution environment that allows separate, unrelated, JavaScript
// code to run in a single instance of IVM.
type Sandbox struct {
	id      int // nolint
	flags   int64
	rt      *quickjs.Runtime
	ctx     *quickjs.Context
	host    *host.Host
	gasUsed int64
}

var sbxMap sync.Map

// NewSandbox generate new sandbox for VM and insert into sandbox map
func NewSandbox(e *VM, flags int64) *Sandbox {
	rt, err := quickjs.NewRuntime()
	if err != nil {
		panic(err)
	}
	rt.SetMemoryLimit(100 * 1024 * 1024) // 100MB
	rt.SetMaxStackSize(2 * 1024 * 1024)  // 2MB - let QuickJS catch stack overflow before WASM crashes

	ctx, err := rt.NewContext()
	if err != nil {
		panic(err)
	}

	s := &Sandbox{
		rt:    rt,
		ctx:   ctx,
		flags: flags,
	}
	// Store in map BEFORE Init so that Go callbacks triggered during
	// library loading can resolve the sandbox via getSbx.
	sbxMap.Store(ctx, s)
	s.Init(e.vmType)
	return s
}

// Release release sandbox and delete from map
func (sbx *Sandbox) Release() {
	if sbx.ctx != nil {
		sbxMap.Delete(sbx.ctx)
		sbx.ctx.Close()
		sbx.ctx = nil
	}
	if sbx.rt != nil {
		sbx.rt.Close()
		sbx.rt = nil
	}
}

// gasIncrReplacer replaces _IOSTInstruction_counter.incr(...) with
// __IOST_internal_gas_acc += ... to avoid thousands of wazero Go callbacks.
// It properly removes the matching closing parenthesis.
func gasIncrReplacer(code string) string {
	const prefix = "_IOSTInstruction_counter.incr("
	var result strings.Builder
	i := 0
	for i < len(code) {
		idx := strings.Index(code[i:], prefix)
		if idx == -1 {
			result.WriteString(code[i:])
			break
		}
		idx += i
		result.WriteString(code[i:idx])
		result.WriteString("__IOST_internal_gas_acc += ((")
		// find matching ')' for the incr( call
		start := idx + len(prefix)
		depth := 1
		j := start
		for j < len(code) && depth > 0 {
			switch code[j] {
			case '(':
				depth++
			case ')':
				depth--
			}
			j++
		}
		// write the argument without the outermost parentheses
		if depth == 0 {
			arg := code[start : j-1]
			result.WriteString(arg)
			result.WriteString(") - (")
			result.WriteString(arg)
			result.WriteString(") % 1)")
		} else {
			// unmatched paren - write everything to end
			arg := code[start:]
			result.WriteString(arg)
			result.WriteString(") - (")
			result.WriteString(arg)
			result.WriteString(") % 1)")
			break
		}
		i = j
	}
	return result.String()
}

// Init add system functions
func (sbx *Sandbox) Init(vmType vmPoolType) {
	ctx := sbx.ctx

	// Inject native helpers
	if vmType == CompileVMPool {
		// Compile VM only needs _cLog and _native_require
		ctx.SetGlobal("_cLog", ctx.Function("_cLog", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
			return ctx.Undefined()
		}))

		// Load compile base libraries in one script (same as old V8 compileCodeFormat)
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
		if _, err := ctx.Eval(compileBootstrap); err != nil {
			panic(fmt.Sprintf("load compile bootstrap error: %v", err))
		}
		return
	}

	// Run VM: inject all host bindings
	ctx.SetGlobal("_cLog", newCLog(ctx))

	// QuickJS ctx.Function cannot be used with 'new', so we inject raw Go
	// helpers with _create suffix and wrap them in JS constructors.
	ctx.SetGlobal("_IOSTBlockchain_create", ctx.Function("_IOSTBlockchain_create", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return newIOSTBlockchain(ctx)
	}))
	ctx.SetGlobal("_IOSTStorage_create", ctx.Function("_IOSTStorage_create", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return newIOSTStorage(ctx)
	}))
	ctx.SetGlobal("_IOSTInstruction_create", ctx.Function("_IOSTInstruction_create", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return newIOSTInstruction(ctx)
	}))
	ctx.SetGlobal("__IOSTCrypto_create", ctx.Function("__IOSTCrypto_create", func(ctx *quickjs.Context, this quickjs.Value, args []quickjs.Value) quickjs.Value {
		return newIOSTCrypto(ctx)
	}))

	if _, err := ctx.Eval(`
function IOSTBlockchain() { return _IOSTBlockchain_create(); }
function IOSTStorage() { return _IOSTStorage_create(); }
function IOSTInstruction() { return _IOSTInstruction_create(); }
function _IOSTCrypto() { return __IOSTCrypto_create(); }
`); err != nil {
		panic(fmt.Sprintf("inject constructor wrappers error: %v", err))
	}

	// Initialize the batched gas accumulator before loading libraries.
	if _, err := ctx.Eval("var __IOST_internal_gas_acc = 0;"); err != nil {
		panic(fmt.Sprintf("init gas accumulator error: %v", err))
	}
	// Verify that a strict-mode IIFE can access the global accumulator.
	if _, err := ctx.Eval("(function(){'use strict'; __IOST_internal_gas_acc += 5;})();"); err != nil {
		panic(fmt.Sprintf("strict mode gas acc test error: %v", err))
	}
	// Load runtime base libraries in one script (same as old V8 codeFormat)
	// Replace incr() calls with batched JS accumulation to avoid wazero
	// callback overhead which can be ~10ms per call.
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
		gasIncrReplacer(jsonJS), gasIncrReplacer(bignumberJS), gasIncrReplacer(int64JS), gasIncrReplacer(float64JS), gasIncrReplacer(utilsJS), gasIncrReplacer(consoleJS),
	)
	if _, err := ctx.Eval(runtimeBootstrap); err != nil {
		panic(fmt.Sprintf("load runtime bootstrap error: %v", err))
	}

	// Pre-load storage and blockchain modules into global cache.
	// We cannot call ctx.Eval from inside a Go callback (wazero reentrancy bug),
	// so we pre-load them here and inject pure-JS _native_require/_native_run.
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
		if _, err := ctx.Eval(jsCode); err != nil {
			panic(fmt.Sprintf("load %s module error: %v", mod.name, err))
		}
	}

	// Inject pure-JS require/run helpers that use the pre-loaded modules.
	if _, err := ctx.Eval(`
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

	// Load vm.js (it will require storage.js and blockchain.js via _native_require/_native_run)
	if _, err := ctx.Eval(gasIncrReplacer(vmJS)); err != nil {
		panic(fmt.Sprintf("load vm.js error: %v", err))
	}

	// Remove _native_run from global scope after vm.js has captured it.
	// The old V8 sandbox only exposed _native_run during Init, not during Execute.
	if _, err := ctx.Eval("_native_run = undefined;"); err != nil {
		panic(fmt.Sprintf("remove _native_run error: %v", err))
	}

	// QuickJS may not define Atomics/Intl/WebAssembly; environment.js
	// assigns them to null in strict mode, which throws if undefined.
	// _native_log is referenced by environment.js but not used in QJS path.
	if _, err := ctx.Eval(`
var _native_log = function(){};
if (typeof Atomics === 'undefined') { var Atomics = {}; }
if (typeof Intl === 'undefined') { var Intl = {}; }
if (typeof WebAssembly === 'undefined') { var WebAssembly = {}; }
`); err != nil {
		panic(fmt.Sprintf("inject missing globals error: %v", err))
	}

	// Capture methods that environment.js sets to null so we can restore them.
	if _, err := ctx.Eval(`
var __savedArrayFrom = Array.from;
var __savedArrayOf = Array.of;
var __savedFunction = Function;
`); err != nil {
		panic(fmt.Sprintf("capture globals error: %v", err))
	}

	// Load environment.js with batched gas replacement.
	if _, err := ctx.Eval(gasIncrReplacer(environmentJS)); err != nil {
		panic(fmt.Sprintf("load environment.js error: %v", err))
	}

	// Restore Array.from and Array.of (the old V8 apparently kept them working).
	if _, err := ctx.Eval(`
Array.from = __savedArrayFrom;
Array.of = __savedArrayOf;
`); err != nil {
		panic(fmt.Sprintf("restore Array methods error: %v", err))
	}

	// Block dangerous dynamic-code features to match old V8 sandbox.
	// environment.js sets Function = null, so we restore it first via the
	// captured reference, then replace it with a blocked version.
	if _, err := ctx.Eval(`
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
	// The old V8 C++ code did this via InitBlockchain/InitStorage after loading
	// environment.js; we must do the same here.
	if _, err := ctx.Eval(`
		IOSTBlockchain = function() { return _IOSTBlockchain_create(); };
		IOSTStorage    = function() { return _IOSTStorage_create(); };
		IOSTInstruction= function() { return _IOSTInstruction_create(); };
		_IOSTCrypto    = function() { return __IOSTCrypto_create(); };
	`); err != nil {
		panic(fmt.Sprintf("re-inject native constructors error: %v", err))
	}
}

// GetFlags ...
func (sbx *Sandbox) GetFlags() int64 {
	return sbx.flags
}

// SetGasLimit set gas limit in context
func (sbx *Sandbox) SetGasLimit(limit int64) {
	sbx.gasUsed = 0
}

// SetHost set host in sandbox and set gas limit
func (sbx *Sandbox) SetHost(host *host.Host) {
	sbx.host = host
	sbx.SetGasLimit(host.GasLimitValue())
}

// Validate contract before save, return err if invalid
func (sbx *Sandbox) Validate(contract *contract.Contract) error {
	code := moduleReplacer.Replace(contract.Code)
	abi, _ := json.Marshal(contract.Info.Abi)

	jsCode := "(function(){\nconst source = \"" + code + "\";\nconst abi = " + string(abi) + ";\nreturn validate(source, abi);\n})();"
	result, err := sbx.ctx.Eval(jsCode)
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

// Compile contract before execution, return compiled code
func (sbx *Sandbox) Compile(contract *contract.Contract) (string, error) {
	code := moduleReplacer.Replace(contract.Code)

	jsCode := "(function(){\nconst source = \"" + code + "\";\nreturn injectGas(source);\n})();"
	result, err := sbx.ctx.Eval(jsCode)
	if err != nil {
		return "", err
	}
	rs := result.String()
	if rs == "" {
		return "", errors.New("inject gas failed")
	}
	return rs, nil
}

// Prepare for contract, inject code
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

// Execute prepared code, return results, gasUsed
func (sbx *Sandbox) Execute(preparedCode string) (string, int64, error) {
	now := time.Now()
	if !sbx.host.Deadline().After(now) {
		return "", 0, errors.New("execution killed")
	}

	// Reset gas used so that Init-time library loading gas is not counted.
	sbx.gasUsed = 0

	// Preload block/tx globals exactly like the old V8 sandbox.cc did.
	// We use JS blockchain.blockInfo()/txInfo() so the JSON string is
	// passed directly to JSON.parse (no JS-string-literal round-trip).
	if _, err := sbx.ctx.Eval(`
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

	// reset gas accumulator for this execution
	if _, err := sbx.ctx.Eval("__IOST_internal_gas_acc = 0;"); err != nil {
		return "", sbx.gasUsed, err
	}
	sbx.gasUsed = 0

	// Transform prepared code to use batched gas accumulation
	preparedCode = gasIncrReplacer(preparedCode)

	// Wrap in IIFE with try-catch so thrown primitives (strings, undefined,
	// etc.) are converted to Error objects.  The QuickJS Go bridge only
	// extracts the "message" property for errors; for primitives it sees
	// "undefined" and loses the real value.
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

	result, err := sbx.ctx.Eval(wrappedCode)

	// Read the batched gas accumulator even if execution failed,
	// because gas is charged for work done before the error.
	gasVal, _ := sbx.ctx.Eval("__IOST_internal_gas_acc;")
	if gasVal.IsNumber() {
		if g, err2 := gasVal.Int64(); err2 == nil {
			sbx.gasUsed += g
		}
	}

	if err != nil {
		return "", sbx.gasUsed, err
	}

	// Match old V8 behavior: undefined -> "", null -> "null", others -> string value
	var str string
	if result.IsUndefined() {
		str = ""
	} else if result.IsNull() {
		str = "null"
	} else {
		str = result.StringBytes()
	}

	if len(str) > resultMaxLength {
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
