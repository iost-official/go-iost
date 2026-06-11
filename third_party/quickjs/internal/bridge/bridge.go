// Package bridge provides low-level bindings to the QuickJS-ng WASM module.
package bridge

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/Gaurav-Gosain/quickjs/wasm"
)

// Global compilation cache - the compilation cache speeds up CompileModule
// by caching the compiled machine code. This is shared across all Bridge instances.
var (
	globalCache     wazero.CompilationCache
	globalCacheOnce sync.Once
)

// initGlobalCache initializes the global compilation cache (called once)
func initGlobalCache() {
	globalCache = wazero.NewCompilationCache()
}

// Buffer pool to reduce allocations for small temporary buffers
var bufPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 256)
		return &buf
	},
}

func getBuffer() *[]byte {
	return bufPool.Get().(*[]byte)
}

func putBuffer(buf *[]byte) {
	*buf = (*buf)[:0]
	bufPool.Put(buf)
}

// GoFunc is a Go function that can be called from JavaScript.
// It receives the context pointer and the arguments as JSValue pointers.
type GoFunc func(ctxPtr uint32, args []uint32) uint32

// Bridge manages the WASM runtime and provides low-level access to QuickJS-ng functions.
type Bridge struct {
	wasmRuntime wazero.Runtime
	module      api.Module
	memory      api.Memory
	mu          sync.Mutex
	logFunc     func(msg string)

	// Go function callbacks
	callbacks  map[uint32]GoFunc // funcID -> Go function
	nextFuncID uint32
	callbackMu sync.RWMutex

	// Exported functions from WASM
	fnAlloc               api.Function
	fnFree                api.Function
	fnGetHeapPtr          api.Function
	fnGetHeapSize         api.Function
	fnResetHeap           api.Function
	fnNewRuntime          api.Function
	fnFreeRuntime         api.Function
	fnNewContext          api.Function
	fnFreeContext         api.Function
	fnGetRuntime          api.Function
	fnEval                api.Function
	fnEvalModule          api.Function
	fnIsException         api.Function
	fnIsUndefined         api.Function
	fnIsNull              api.Function
	fnIsBool              api.Function
	fnIsNumber            api.Function
	fnIsString            api.Function
	fnIsSymbol            api.Function
	fnIsObject            api.Function
	fnIsFunction          api.Function
	fnIsArray             api.Function
	fnIsError             api.Function
	fnIsBigInt            api.Function
	fnIsDate              api.Function
	fnIsRegExp            api.Function
	fnIsMap               api.Function
	fnIsSet               api.Function
	fnToBool              api.Function
	fnToInt32             api.Function
	fnToInt64             api.Function
	fnToFloat64           api.Function
	fnToCString           api.Function
	fnFreeCString         api.Function
	fnToCStringLen        api.Function
	fnNewUndefined        api.Function
	fnNewNull             api.Function
	fnNewBool             api.Function
	fnNewInt32            api.Function
	fnNewInt64            api.Function
	fnNewFloat64          api.Function
	fnNewString           api.Function
	fnNewStringLen        api.Function
	fnNewObject           api.Function
	fnNewArray            api.Function
	fnGetProperty         api.Function
	fnSetProperty         api.Function
	fnHasProperty         api.Function
	fnDeleteProperty      api.Function
	fnGetPropertyUint32   api.Function
	fnSetPropertyUint32   api.Function
	fnGetGlobalObject     api.Function
	fnCall                api.Function
	fnCallConstructor     api.Function
	fnInvoke              api.Function
	fnGetException        api.Function
	fnHasException        api.Function
	fnThrow               api.Function
	fnThrowError          api.Function
	fnThrowTypeError      api.Function
	fnThrowRangeError     api.Function
	fnThrowSyntaxError    api.Function
	fnThrowReferenceError api.Function
	fnDupValue            api.Function
	fnFreeValue           api.Function
	fnJSONParse           api.Function
	fnJSONStringify       api.Function
	fnRunGC               api.Function
	fnIsPromise           api.Function
	fnNewPromise          api.Function
	fnExecutePendingJobs  api.Function
	fnNewBigInt64         api.Function
	fnNewBigUint64        api.Function
	fnToBigInt64          api.Function
	fnNewDate             api.Function
	fnInstanceof          api.Function
	fnTypeof              api.Function
	fnGetOwnPropertyNames api.Function
	fnNewArrayBuffer      api.Function
	fnGetArrayBuffer      api.Function
	fnStdAddConsole       api.Function
	fnNewCFunction        api.Function
	fnStrictEq            api.Function
	fnSetMemoryLimit      api.Function
	fnSetMaxStackSize     api.Function
	fnGetErrorMessage     api.Function
	fnGetErrorStack       api.Function
	fnToString            api.Function
}

// New creates a new Bridge instance.
func New(ctx context.Context) (*Bridge, error) {
	b := &Bridge{
		logFunc: func(msg string) {
			fmt.Print(msg)
		},
		callbacks:  make(map[uint32]GoFunc),
		nextFuncID: 1,
	}

	// Initialize global compilation cache once
	globalCacheOnce.Do(initGlobalCache)

	// Create optimized wazero runtime config:
	// - Use compilation cache to speed up CompileModule (caches compiled machine code)
	// - Disable debug info for faster execution (no DWARF parsing)
	runtimeConfig := wazero.NewRuntimeConfig().
		WithCompilationCache(globalCache).
		WithDebugInfoEnabled(false)

	b.wasmRuntime = wazero.NewRuntimeWithConfig(ctx, runtimeConfig)

	// Instantiate WASI - required by the QuickJS WASM module
	wasi_snapshot_preview1.MustInstantiate(ctx, b.wasmRuntime)

	// Register host functions
	_, err := b.wasmRuntime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(b.hostLog).
		Export("host_log").
		NewFunctionBuilder().
		WithFunc(b.hostCallGo).
		Export("host_call_go").
		Instantiate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate host module: %w", err)
	}

	// Compile the WASM module - the compilation cache makes subsequent compiles fast
	compiled, err := b.wasmRuntime.CompileModule(ctx, wasm.QuickJS)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	b.module, err = b.wasmRuntime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	b.memory = b.module.Memory()
	if b.memory == nil {
		return nil, errors.New("WASM module has no memory")
	}

	// Get all exported functions
	if err := b.initFunctions(); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bridge) initFunctions() error {
	getFn := func(name string) (api.Function, error) {
		fn := b.module.ExportedFunction(name)
		if fn == nil {
			return nil, fmt.Errorf("function %s not found in WASM module", name)
		}
		return fn, nil
	}

	var err error

	// Memory management
	if b.fnAlloc, err = getFn("qjs_alloc"); err != nil {
		return err
	}
	if b.fnFree, err = getFn("qjs_free"); err != nil {
		return err
	}
	if b.fnGetHeapPtr, err = getFn("qjs_get_heap_ptr"); err != nil {
		return err
	}
	if b.fnGetHeapSize, err = getFn("qjs_get_heap_size"); err != nil {
		return err
	}
	if b.fnResetHeap, err = getFn("qjs_reset_heap"); err != nil {
		return err
	}

	// Runtime and context
	if b.fnNewRuntime, err = getFn("qjs_new_runtime"); err != nil {
		return err
	}
	if b.fnFreeRuntime, err = getFn("qjs_free_runtime"); err != nil {
		return err
	}
	if b.fnNewContext, err = getFn("qjs_new_context"); err != nil {
		return err
	}
	if b.fnFreeContext, err = getFn("qjs_free_context"); err != nil {
		return err
	}
	if b.fnGetRuntime, err = getFn("qjs_get_runtime"); err != nil {
		return err
	}

	// Evaluation
	if b.fnEval, err = getFn("qjs_eval"); err != nil {
		return err
	}
	if b.fnEvalModule, err = getFn("qjs_eval_module"); err != nil {
		return err
	}

	// Type checking
	if b.fnIsException, err = getFn("qjs_is_exception"); err != nil {
		return err
	}
	if b.fnIsUndefined, err = getFn("qjs_is_undefined"); err != nil {
		return err
	}
	if b.fnIsNull, err = getFn("qjs_is_null"); err != nil {
		return err
	}
	if b.fnIsBool, err = getFn("qjs_is_bool"); err != nil {
		return err
	}
	if b.fnIsNumber, err = getFn("qjs_is_number"); err != nil {
		return err
	}
	if b.fnIsString, err = getFn("qjs_is_string"); err != nil {
		return err
	}
	if b.fnIsSymbol, err = getFn("qjs_is_symbol"); err != nil {
		return err
	}
	if b.fnIsObject, err = getFn("qjs_is_object"); err != nil {
		return err
	}
	if b.fnIsFunction, err = getFn("qjs_is_function"); err != nil {
		return err
	}
	if b.fnIsArray, err = getFn("qjs_is_array"); err != nil {
		return err
	}
	if b.fnIsError, err = getFn("qjs_is_error"); err != nil {
		return err
	}
	if b.fnIsBigInt, err = getFn("qjs_is_big_int"); err != nil {
		return err
	}
	if b.fnIsDate, err = getFn("qjs_is_date"); err != nil {
		return err
	}
	if b.fnIsRegExp, err = getFn("qjs_is_regexp"); err != nil {
		return err
	}
	if b.fnIsMap, err = getFn("qjs_is_map"); err != nil {
		return err
	}
	if b.fnIsSet, err = getFn("qjs_is_set"); err != nil {
		return err
	}

	// Value conversion
	if b.fnToBool, err = getFn("qjs_to_bool"); err != nil {
		return err
	}
	if b.fnToInt32, err = getFn("qjs_to_int32"); err != nil {
		return err
	}
	if b.fnToInt64, err = getFn("qjs_to_int64"); err != nil {
		return err
	}
	if b.fnToFloat64, err = getFn("qjs_to_float64"); err != nil {
		return err
	}
	if b.fnToCString, err = getFn("qjs_to_cstring"); err != nil {
		return err
	}
	if b.fnFreeCString, err = getFn("qjs_free_cstring"); err != nil {
		return err
	}
	if b.fnToCStringLen, err = getFn("qjs_to_cstring_len"); err != nil {
		return err
	}

	// Value creation
	if b.fnNewUndefined, err = getFn("qjs_new_undefined"); err != nil {
		return err
	}
	if b.fnNewNull, err = getFn("qjs_new_null"); err != nil {
		return err
	}
	if b.fnNewBool, err = getFn("qjs_new_bool"); err != nil {
		return err
	}
	if b.fnNewInt32, err = getFn("qjs_new_int32"); err != nil {
		return err
	}
	if b.fnNewInt64, err = getFn("qjs_new_int64"); err != nil {
		return err
	}
	if b.fnNewFloat64, err = getFn("qjs_new_float64"); err != nil {
		return err
	}
	if b.fnNewString, err = getFn("qjs_new_string"); err != nil {
		return err
	}
	if b.fnNewStringLen, err = getFn("qjs_new_string_len"); err != nil {
		return err
	}

	// Object operations
	if b.fnNewObject, err = getFn("qjs_new_object"); err != nil {
		return err
	}
	if b.fnNewArray, err = getFn("qjs_new_array"); err != nil {
		return err
	}
	if b.fnGetProperty, err = getFn("qjs_get_property"); err != nil {
		return err
	}
	if b.fnSetProperty, err = getFn("qjs_set_property"); err != nil {
		return err
	}
	if b.fnHasProperty, err = getFn("qjs_has_property"); err != nil {
		return err
	}
	if b.fnDeleteProperty, err = getFn("qjs_delete_property"); err != nil {
		return err
	}
	if b.fnGetPropertyUint32, err = getFn("qjs_get_property_uint32"); err != nil {
		return err
	}
	if b.fnSetPropertyUint32, err = getFn("qjs_set_property_uint32"); err != nil {
		return err
	}
	if b.fnGetGlobalObject, err = getFn("qjs_get_global_object"); err != nil {
		return err
	}

	// Function calling
	if b.fnCall, err = getFn("qjs_call"); err != nil {
		return err
	}
	if b.fnCallConstructor, err = getFn("qjs_call_constructor"); err != nil {
		return err
	}
	if b.fnInvoke, err = getFn("qjs_invoke"); err != nil {
		return err
	}

	// Exception handling
	if b.fnGetException, err = getFn("qjs_get_exception"); err != nil {
		return err
	}
	if b.fnHasException, err = getFn("qjs_has_exception"); err != nil {
		return err
	}
	if b.fnThrow, err = getFn("qjs_throw"); err != nil {
		return err
	}
	if b.fnThrowError, err = getFn("qjs_throw_error"); err != nil {
		return err
	}
	if b.fnThrowTypeError, err = getFn("qjs_throw_type_error"); err != nil {
		return err
	}
	if b.fnThrowRangeError, err = getFn("qjs_throw_range_error"); err != nil {
		return err
	}
	if b.fnThrowSyntaxError, err = getFn("qjs_throw_syntax_error"); err != nil {
		return err
	}
	if b.fnThrowReferenceError, err = getFn("qjs_throw_reference_error"); err != nil {
		return err
	}

	// Value management
	if b.fnDupValue, err = getFn("qjs_dup_value"); err != nil {
		return err
	}
	if b.fnFreeValue, err = getFn("qjs_free_value"); err != nil {
		return err
	}

	// JSON
	if b.fnJSONParse, err = getFn("qjs_json_parse"); err != nil {
		return err
	}
	if b.fnJSONStringify, err = getFn("qjs_json_stringify"); err != nil {
		return err
	}

	// GC
	if b.fnRunGC, err = getFn("qjs_run_gc"); err != nil {
		return err
	}

	// Promise
	if b.fnIsPromise, err = getFn("qjs_is_promise"); err != nil {
		return err
	}
	if b.fnNewPromise, err = getFn("qjs_new_promise"); err != nil {
		return err
	}
	if b.fnExecutePendingJobs, err = getFn("qjs_execute_pending_jobs"); err != nil {
		return err
	}

	// BigInt
	if b.fnNewBigInt64, err = getFn("qjs_new_big_int64"); err != nil {
		return err
	}
	if b.fnNewBigUint64, err = getFn("qjs_new_big_uint64"); err != nil {
		return err
	}
	if b.fnToBigInt64, err = getFn("qjs_to_big_int64"); err != nil {
		return err
	}

	// Date
	if b.fnNewDate, err = getFn("qjs_new_date"); err != nil {
		return err
	}

	// Type operations
	if b.fnInstanceof, err = getFn("qjs_instanceof"); err != nil {
		return err
	}
	if b.fnTypeof, err = getFn("qjs_typeof"); err != nil {
		return err
	}

	// Property enumeration
	if b.fnGetOwnPropertyNames, err = getFn("qjs_get_own_property_names"); err != nil {
		return err
	}

	// ArrayBuffer
	if b.fnNewArrayBuffer, err = getFn("qjs_new_array_buffer"); err != nil {
		return err
	}
	if b.fnGetArrayBuffer, err = getFn("qjs_get_array_buffer"); err != nil {
		return err
	}

	// Console
	if b.fnStdAddConsole, err = getFn("qjs_std_add_console"); err != nil {
		return err
	}

	// C function binding
	if b.fnNewCFunction, err = getFn("qjs_new_c_function"); err != nil {
		return err
	}

	// Equality
	if b.fnStrictEq, err = getFn("qjs_strict_eq"); err != nil {
		return err
	}

	// Runtime configuration
	if b.fnSetMemoryLimit, err = getFn("qjs_set_memory_limit"); err != nil {
		return err
	}
	if b.fnSetMaxStackSize, err = getFn("qjs_set_max_stack_size"); err != nil {
		return err
	}

	// Error utilities
	if b.fnGetErrorMessage, err = getFn("qjs_get_error_message"); err != nil {
		return err
	}
	if b.fnGetErrorStack, err = getFn("qjs_get_error_stack"); err != nil {
		return err
	}

	// String conversion
	if b.fnToString, err = getFn("qjs_to_string"); err != nil {
		return err
	}

	return nil
}

// Close releases all resources.
func (b *Bridge) Close(ctx context.Context) error {
	return b.wasmRuntime.Close(ctx)
}

// SetLogFunc sets the function used for console output from JavaScript.
func (b *Bridge) SetLogFunc(fn func(msg string)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.logFunc = fn
}

// Host function implementations

func (b *Bridge) hostLog(ctx context.Context, m api.Module, bufPtr, bufLen uint32) {
	buf, ok := m.Memory().Read(bufPtr, bufLen)
	if !ok {
		return
	}
	b.mu.Lock()
	logFunc := b.logFunc
	b.mu.Unlock()
	if logFunc != nil {
		logFunc(string(buf))
	}
}

func (b *Bridge) hostCallGo(ctx context.Context, m api.Module, ctxPtr, funcID uint32, argc int32, argvPtr uint32) uint32 {
	b.callbackMu.RLock()
	fn, ok := b.callbacks[funcID]
	b.callbackMu.RUnlock()

	if !ok {
		// Function not found, return undefined
		undef, _ := b.NewUndefined(ctx)
		return undef
	}

	// Read argument pointers from WASM memory
	args := make([]uint32, argc)
	if argc > 0 && argvPtr != 0 {
		for i := range argc {
			buf, ok := m.Memory().Read(argvPtr+uint32(i)*4, 4)
			if !ok {
				undef, _ := b.NewUndefined(ctx)
				return undef
			}
			args[i] = binary.LittleEndian.Uint32(buf)
		}
	}

	// Call the Go function
	return fn(ctxPtr, args)
}

// Memory management helpers

// Alloc allocates memory in WASM heap and returns the pointer.
func (b *Bridge) Alloc(ctx context.Context, size uint32) (uint32, error) {
	results, err := b.fnAlloc.Call(ctx, uint64(size))
	if err != nil {
		return 0, err
	}
	ptr := uint32(results[0])
	if ptr == 0 {
		return 0, errors.New("WASM allocation failed")
	}
	return ptr, nil
}

// Free frees memory in WASM heap.
func (b *Bridge) Free(ctx context.Context, ptr uint32) error {
	_, err := b.fnFree.Call(ctx, uint64(ptr))
	return err
}

// WriteString writes a string to WASM memory and returns the pointer.
func (b *Bridge) WriteString(ctx context.Context, s string) (ptr uint32, err error) {
	sLen := len(s)
	ptr, err = b.Alloc(ctx, uint32(sLen+1)) // +1 for null terminator
	if err != nil {
		return 0, err
	}

	// For small strings, use pooled buffer to avoid allocation
	if sLen < 256 {
		bufPtr := getBuffer()
		*bufPtr = append((*bufPtr)[:0], s...)
		*bufPtr = append(*bufPtr, 0)
		ok := b.memory.Write(ptr, *bufPtr)
		putBuffer(bufPtr)
		if !ok {
			return 0, errors.New("failed to write string to WASM memory")
		}
	} else {
		// For larger strings, allocate directly
		data := make([]byte, sLen+1)
		copy(data, s)
		data[sLen] = 0
		if !b.memory.Write(ptr, data) {
			return 0, errors.New("failed to write string to WASM memory")
		}
	}
	return ptr, nil
}

// WriteBytes writes bytes to WASM memory and returns the pointer.
func (b *Bridge) WriteBytes(ctx context.Context, data []byte) (ptr uint32, err error) {
	ptr, err = b.Alloc(ctx, uint32(len(data)))
	if err != nil {
		return 0, err
	}
	if !b.memory.Write(ptr, data) {
		return 0, errors.New("failed to write bytes to WASM memory")
	}
	return ptr, nil
}

// ReadCString reads a null-terminated string from WASM memory.
func (b *Bridge) ReadCString(ptr, maxLen uint32) string {
	buf, ok := b.memory.Read(ptr, maxLen)
	if !ok {
		return ""
	}
	// Use optimized bytes.IndexByte (has assembly implementation)
	if idx := bytes.IndexByte(buf, 0); idx >= 0 {
		return string(buf[:idx])
	}
	return string(buf)
}

// ReadBytes reads bytes from WASM memory.
func (b *Bridge) ReadBytes(ptr, length uint32) []byte {
	buf, ok := b.memory.Read(ptr, length)
	if !ok {
		return nil
	}
	result := make([]byte, length)
	copy(result, buf)
	return result
}

// Memory returns the WASM memory for direct access.
func (b *Bridge) Memory() api.Memory {
	return b.memory
}

// ============================================================================
// Runtime and Context Management
// ============================================================================

// NewRuntime creates a new JavaScript runtime.
func (b *Bridge) NewRuntime(ctx context.Context) (uint32, error) {
	results, err := b.fnNewRuntime.Call(ctx)
	if err != nil {
		return 0, err
	}
	rtPtr := uint32(results[0])
	if rtPtr == 0 {
		return 0, errors.New("failed to create JavaScript runtime")
	}
	return rtPtr, nil
}

// FreeRuntime frees a JavaScript runtime.
func (b *Bridge) FreeRuntime(ctx context.Context, rtPtr uint32) error {
	_, err := b.fnFreeRuntime.Call(ctx, uint64(rtPtr))
	return err
}

// NewContext creates a new JavaScript context.
func (b *Bridge) NewContext(ctx context.Context, rtPtr uint32) (uint32, error) {
	results, err := b.fnNewContext.Call(ctx, uint64(rtPtr))
	if err != nil {
		return 0, err
	}
	ctxPtr := uint32(results[0])
	if ctxPtr == 0 {
		return 0, errors.New("failed to create JavaScript context")
	}
	return ctxPtr, nil
}

// FreeContext frees a JavaScript context.
func (b *Bridge) FreeContext(ctx context.Context, ctxPtr uint32) error {
	_, err := b.fnFreeContext.Call(ctx, uint64(ctxPtr))
	return err
}

// GetRuntime gets the runtime from a context.
func (b *Bridge) GetRuntime(ctx context.Context, ctxPtr uint32) (uint32, error) {
	results, err := b.fnGetRuntime.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AddConsole adds console.log and print functions to the context.
func (b *Bridge) AddConsole(ctx context.Context, ctxPtr uint32) error {
	_, err := b.fnStdAddConsole.Call(ctx, uint64(ctxPtr))
	return err
}

// ============================================================================
// Evaluation
// ============================================================================

// Eval evaluates JavaScript code.
func (b *Bridge) Eval(ctx context.Context, ctxPtr uint32, code, filename string, flags int32) (uint32, error) {
	codePtr, err := b.WriteString(ctx, code)
	if err != nil {
		return 0, err
	}

	var filenamePtr uint32
	if filename != "" {
		filenamePtr, err = b.WriteString(ctx, filename)
		if err != nil {
			return 0, err
		}
	}

	results, err := b.fnEval.Call(ctx, uint64(ctxPtr), uint64(codePtr), uint64(len(code)), uint64(filenamePtr), uint64(flags))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// EvalModule evaluates JavaScript module code.
func (b *Bridge) EvalModule(ctx context.Context, ctxPtr uint32, code, filename string) (uint32, error) {
	codePtr, err := b.WriteString(ctx, code)
	if err != nil {
		return 0, err
	}

	var filenamePtr uint32
	if filename != "" {
		filenamePtr, err = b.WriteString(ctx, filename)
		if err != nil {
			return 0, err
		}
	}

	results, err := b.fnEvalModule.Call(ctx, uint64(ctxPtr), uint64(codePtr), uint64(len(code)), uint64(filenamePtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Type Checking
// ============================================================================

func (b *Bridge) IsException(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsException.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsUndefined(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsUndefined.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsNull(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsNull.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsBool(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsBool.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsNumber(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsNumber.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsString(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsString.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsSymbol(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsSymbol.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsObject(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsObject.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsFunction(ctx context.Context, ctxPtr, valPtr uint32) (bool, error) {
	results, err := b.fnIsFunction.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsArray(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsArray.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsError(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsError.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsBigInt(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsBigInt.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsDate(ctx context.Context, valPtr uint32) (bool, error) {
	results, err := b.fnIsDate.Call(ctx, uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) IsPromise(ctx context.Context, ctxPtr, valPtr uint32) (bool, error) {
	results, err := b.fnIsPromise.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

// ============================================================================
// Value Conversion
// ============================================================================

func (b *Bridge) ToBool(ctx context.Context, ctxPtr, valPtr uint32) (bool, error) {
	results, err := b.fnToBool.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return false, err
	}
	return int32(results[0]) > 0, nil
}

func (b *Bridge) ToInt32(ctx context.Context, ctxPtr, valPtr uint32) (int32, error) {
	resultPtr, err := b.Alloc(ctx, 4)
	if err != nil {
		return 0, err
	}

	results, err := b.fnToInt32.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(resultPtr))
	if err != nil {
		return 0, err
	}
	if int32(results[0]) != 0 {
		return 0, errors.New("ToInt32 conversion failed")
	}

	buf, ok := b.memory.Read(resultPtr, 4)
	if !ok {
		return 0, errors.New("failed to read result from WASM memory")
	}
	return int32(binary.LittleEndian.Uint32(buf)), nil
}

func (b *Bridge) ToInt64(ctx context.Context, ctxPtr, valPtr uint32) (int64, error) {
	resultPtr, err := b.Alloc(ctx, 8)
	if err != nil {
		return 0, err
	}

	results, err := b.fnToInt64.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(resultPtr))
	if err != nil {
		return 0, err
	}
	if int32(results[0]) != 0 {
		return 0, errors.New("ToInt64 conversion failed")
	}

	buf, ok := b.memory.Read(resultPtr, 8)
	if !ok {
		return 0, errors.New("failed to read result from WASM memory")
	}
	return int64(binary.LittleEndian.Uint64(buf)), nil
}

func (b *Bridge) ToFloat64(ctx context.Context, ctxPtr, valPtr uint32) (float64, error) {
	resultPtr, err := b.Alloc(ctx, 8)
	if err != nil {
		return 0, err
	}

	results, err := b.fnToFloat64.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(resultPtr))
	if err != nil {
		return 0, err
	}
	if int32(results[0]) != 0 {
		return 0, errors.New("ToFloat64 conversion failed")
	}

	buf, ok := b.memory.Read(resultPtr, 8)
	if !ok {
		return 0, errors.New("failed to read result from WASM memory")
	}
	bits := binary.LittleEndian.Uint64(buf)
	return math.Float64frombits(bits), nil
}

func (b *Bridge) ToString(ctx context.Context, ctxPtr, valPtr uint32) (string, error) {
	// Get the string using JS_ToCString
	results, err := b.fnToCString.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return "", err
	}
	strPtr := uint32(results[0])
	if strPtr == 0 {
		return "", nil
	}

	// Read the string (up to 64KB)
	str := b.ReadCString(strPtr, 65536)

	// Free the C string
	_, _ = b.fnFreeCString.Call(ctx, uint64(ctxPtr), uint64(strPtr))

	return str, nil
}

// ToStringLen returns the string representation of a JS value, preserving null bytes.
func (b *Bridge) ToStringLen(ctx context.Context, ctxPtr, valPtr uint32) (string, error) {
	// Allocate memory for size_t (4 bytes on WASM32)
	lenPtr, err := b.Alloc(ctx, 4)
	if err != nil {
		return "", err
	}
	defer b.Free(ctx, lenPtr)

	// Get the string using JS_ToCStringLen
	results, err := b.fnToCStringLen.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(lenPtr))
	if err != nil {
		return "", err
	}
	strPtr := uint32(results[0])
	if strPtr == 0 {
		return "", nil
	}

	// Read the length
	lenBuf := b.ReadBytes(lenPtr, 4)
	if lenBuf == nil {
		_, _ = b.fnFreeCString.Call(ctx, uint64(ctxPtr), uint64(strPtr))
		return "", errors.New("failed to read string length")
	}
	strLen := binary.LittleEndian.Uint32(lenBuf)

	// Read the string bytes
	var str string
	if strLen > 0 {
		buf := b.ReadBytes(strPtr, strLen)
		if buf == nil {
			_, _ = b.fnFreeCString.Call(ctx, uint64(ctxPtr), uint64(strPtr))
			return "", errors.New("failed to read string bytes")
		}
		str = string(buf)
	}

	// Free the C string
	_, _ = b.fnFreeCString.Call(ctx, uint64(ctxPtr), uint64(strPtr))

	return str, nil
}

// ============================================================================
// Value Creation
// ============================================================================

func (b *Bridge) NewUndefined(ctx context.Context) (uint32, error) {
	results, err := b.fnNewUndefined.Call(ctx)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewNull(ctx context.Context) (uint32, error) {
	results, err := b.fnNewNull.Call(ctx)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewBool(ctx context.Context, val bool) (uint32, error) {
	v := int32(0)
	if val {
		v = 1
	}
	results, err := b.fnNewBool.Call(ctx, uint64(v))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewInt32(ctx context.Context, val int32) (uint32, error) {
	results, err := b.fnNewInt32.Call(ctx, uint64(val))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewInt64(ctx context.Context, ctxPtr uint32, val int64) (uint32, error) {
	results, err := b.fnNewInt64.Call(ctx, uint64(ctxPtr), uint64(val))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewFloat64(ctx context.Context, val float64) (uint32, error) {
	bits := math.Float64bits(val)
	results, err := b.fnNewFloat64.Call(ctx, bits)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewString(ctx context.Context, ctxPtr uint32, s string) (uint32, error) {
	strPtr, err := b.WriteString(ctx, s)
	if err != nil {
		return 0, err
	}
	results, err := b.fnNewString.Call(ctx, uint64(ctxPtr), uint64(strPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewStringLen(ctx context.Context, ctxPtr uint32, s string) (uint32, error) {
	strPtr, err := b.WriteString(ctx, s)
	if err != nil {
		return 0, err
	}
	results, err := b.fnNewStringLen.Call(ctx, uint64(ctxPtr), uint64(strPtr), uint64(len(s)))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Object Operations
// ============================================================================

func (b *Bridge) NewObject(ctx context.Context, ctxPtr uint32) (uint32, error) {
	results, err := b.fnNewObject.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewArray(ctx context.Context, ctxPtr uint32) (uint32, error) {
	results, err := b.fnNewArray.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) GetProperty(ctx context.Context, ctxPtr, objPtr uint32, prop string) (uint32, error) {
	propPtr, err := b.WriteString(ctx, prop)
	if err != nil {
		return 0, err
	}
	results, err := b.fnGetProperty.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(propPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) SetProperty(ctx context.Context, ctxPtr, objPtr uint32, prop string, valPtr uint32) error {
	propPtr, err := b.WriteString(ctx, prop)
	if err != nil {
		return err
	}
	results, err := b.fnSetProperty.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(propPtr), uint64(valPtr))
	if err != nil {
		return err
	}
	if int32(results[0]) < 0 {
		return errors.New("failed to set property")
	}
	return nil
}

func (b *Bridge) HasProperty(ctx context.Context, ctxPtr, objPtr uint32, prop string) (bool, error) {
	propPtr, err := b.WriteString(ctx, prop)
	if err != nil {
		return false, err
	}
	results, err := b.fnHasProperty.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(propPtr))
	if err != nil {
		return false, err
	}
	return int32(results[0]) > 0, nil
}

func (b *Bridge) DeleteProperty(ctx context.Context, ctxPtr, objPtr uint32, prop string) error {
	propPtr, err := b.WriteString(ctx, prop)
	if err != nil {
		return err
	}
	results, err := b.fnDeleteProperty.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(propPtr))
	if err != nil {
		return err
	}
	if int32(results[0]) < 0 {
		return errors.New("failed to delete property")
	}
	return nil
}

func (b *Bridge) GetPropertyUint32(ctx context.Context, ctxPtr, objPtr, idx uint32) (uint32, error) {
	results, err := b.fnGetPropertyUint32.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(idx))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) SetPropertyUint32(ctx context.Context, ctxPtr, objPtr, idx, valPtr uint32) error {
	results, err := b.fnSetPropertyUint32.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(idx), uint64(valPtr))
	if err != nil {
		return err
	}
	if int32(results[0]) < 0 {
		return errors.New("failed to set property by index")
	}
	return nil
}

func (b *Bridge) GetGlobalObject(ctx context.Context, ctxPtr uint32) (uint32, error) {
	results, err := b.fnGetGlobalObject.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Function Calling
// ============================================================================

func (b *Bridge) Call(ctx context.Context, ctxPtr, funcPtr, thisPtr uint32, args []uint32) (uint32, error) {
	argc := int32(len(args))
	var argvPtr uint32

	if argc > 0 {
		var err error
		argvPtr, err = b.Alloc(ctx, uint32(argc)*4)
		if err != nil {
			return 0, err
		}
		// Write all argument pointers in one batch for better performance
		argBuf := make([]byte, argc*4)
		for i, arg := range args {
			binary.LittleEndian.PutUint32(argBuf[i*4:], arg)
		}
		if !b.memory.Write(argvPtr, argBuf) {
			return 0, errors.New("failed to write arguments to WASM memory")
		}
	}

	results, err := b.fnCall.Call(ctx, uint64(ctxPtr), uint64(funcPtr), uint64(thisPtr), uint64(argc), uint64(argvPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) CallConstructor(ctx context.Context, ctxPtr, funcPtr uint32, args []uint32) (uint32, error) {
	argc := int32(len(args))
	var argvPtr uint32

	if argc > 0 {
		var err error
		argvPtr, err = b.Alloc(ctx, uint32(argc)*4)
		if err != nil {
			return 0, err
		}
		// Write all argument pointers in one batch
		argBuf := make([]byte, argc*4)
		for i, arg := range args {
			binary.LittleEndian.PutUint32(argBuf[i*4:], arg)
		}
		if !b.memory.Write(argvPtr, argBuf) {
			return 0, errors.New("failed to write arguments to WASM memory")
		}
	}

	results, err := b.fnCallConstructor.Call(ctx, uint64(ctxPtr), uint64(funcPtr), uint64(argc), uint64(argvPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) Invoke(ctx context.Context, ctxPtr, objPtr uint32, method string, args []uint32) (uint32, error) {
	methodPtr, err := b.WriteString(ctx, method)
	if err != nil {
		return 0, err
	}

	argc := int32(len(args))
	var argvPtr uint32

	if argc > 0 {
		argvPtr, err = b.Alloc(ctx, uint32(argc)*4)
		if err != nil {
			return 0, err
		}
		// Write all argument pointers in one batch
		argBuf := make([]byte, argc*4)
		for i, arg := range args {
			binary.LittleEndian.PutUint32(argBuf[i*4:], arg)
		}
		if !b.memory.Write(argvPtr, argBuf) {
			return 0, errors.New("failed to write arguments to WASM memory")
		}
	}

	results, err := b.fnInvoke.Call(ctx, uint64(ctxPtr), uint64(objPtr), uint64(methodPtr), uint64(argc), uint64(argvPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Exception Handling
// ============================================================================

func (b *Bridge) GetException(ctx context.Context, ctxPtr uint32) (uint32, error) {
	results, err := b.fnGetException.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) HasException(ctx context.Context, ctxPtr uint32) (bool, error) {
	results, err := b.fnHasException.Call(ctx, uint64(ctxPtr))
	if err != nil {
		return false, err
	}
	return results[0] != 0, nil
}

func (b *Bridge) ThrowError(ctx context.Context, ctxPtr uint32, msg string) (uint32, error) {
	msgPtr, err := b.WriteString(ctx, msg)
	if err != nil {
		return 0, err
	}
	results, err := b.fnThrowError.Call(ctx, uint64(ctxPtr), uint64(msgPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) ThrowTypeError(ctx context.Context, ctxPtr uint32, msg string) (uint32, error) {
	msgPtr, err := b.WriteString(ctx, msg)
	if err != nil {
		return 0, err
	}
	results, err := b.fnThrowTypeError.Call(ctx, uint64(ctxPtr), uint64(msgPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) GetErrorMessage(ctx context.Context, ctxPtr, errPtr uint32) (string, error) {
	bufPtr, err := b.Alloc(ctx, 1024)
	if err != nil {
		return "", err
	}

	results, err := b.fnGetErrorMessage.Call(ctx, uint64(ctxPtr), uint64(errPtr), uint64(bufPtr), 1024)
	if err != nil {
		return "", err
	}
	msgLen := uint32(results[0])

	return b.ReadCString(bufPtr, msgLen+1), nil
}

func (b *Bridge) GetErrorStack(ctx context.Context, ctxPtr, errPtr uint32) (string, error) {
	results, err := b.fnGetErrorStack.Call(ctx, uint64(ctxPtr), uint64(errPtr))
	if err != nil {
		return "", err
	}
	stackPtr := uint32(results[0])
	return b.ToString(ctx, ctxPtr, stackPtr)
}

// ============================================================================
// Value Management
// ============================================================================

func (b *Bridge) DupValue(ctx context.Context, ctxPtr, valPtr uint32) (uint32, error) {
	results, err := b.fnDupValue.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) FreeValue(ctx context.Context, ctxPtr, valPtr uint32) error {
	_, err := b.fnFreeValue.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	return err
}

// ============================================================================
// JSON
// ============================================================================

func (b *Bridge) JSONParse(ctx context.Context, ctxPtr uint32, json string) (uint32, error) {
	jsonPtr, err := b.WriteString(ctx, json)
	if err != nil {
		return 0, err
	}
	results, err := b.fnJSONParse.Call(ctx, uint64(ctxPtr), uint64(jsonPtr), uint64(len(json)))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) JSONStringify(ctx context.Context, ctxPtr, valPtr uint32) (string, error) {
	results, err := b.fnJSONStringify.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return "", err
	}
	strValPtr := uint32(results[0])
	return b.ToString(ctx, ctxPtr, strValPtr)
}

// ============================================================================
// Garbage Collection
// ============================================================================

func (b *Bridge) RunGC(ctx context.Context, rtPtr uint32) error {
	_, err := b.fnRunGC.Call(ctx, uint64(rtPtr))
	return err
}

// ============================================================================
// Promise Handling
// ============================================================================

func (b *Bridge) ExecutePendingJobs(ctx context.Context, rtPtr uint32) (int32, error) {
	results, err := b.fnExecutePendingJobs.Call(ctx, uint64(rtPtr))
	if err != nil {
		return -1, err
	}
	return int32(results[0]), nil
}

// ============================================================================
// BigInt
// ============================================================================

func (b *Bridge) NewBigInt64(ctx context.Context, ctxPtr uint32, val int64) (uint32, error) {
	results, err := b.fnNewBigInt64.Call(ctx, uint64(ctxPtr), uint64(val))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) NewBigUint64(ctx context.Context, ctxPtr uint32, val uint64) (uint32, error) {
	results, err := b.fnNewBigUint64.Call(ctx, uint64(ctxPtr), val)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) ToBigInt64(ctx context.Context, ctxPtr, valPtr uint32) (int64, error) {
	resultPtr, err := b.Alloc(ctx, 8)
	if err != nil {
		return 0, err
	}

	results, err := b.fnToBigInt64.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(resultPtr))
	if err != nil {
		return 0, err
	}
	if int32(results[0]) != 0 {
		return 0, errors.New("ToBigInt64 conversion failed")
	}

	buf, ok := b.memory.Read(resultPtr, 8)
	if !ok {
		return 0, errors.New("failed to read result from WASM memory")
	}
	return int64(binary.LittleEndian.Uint64(buf)), nil
}

// ============================================================================
// Date
// ============================================================================

func (b *Bridge) NewDate(ctx context.Context, ctxPtr uint32, epochMs float64) (uint32, error) {
	bits := math.Float64bits(epochMs)
	results, err := b.fnNewDate.Call(ctx, uint64(ctxPtr), bits)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Type Operations
// ============================================================================

func (b *Bridge) Instanceof(ctx context.Context, ctxPtr, valPtr, ctorPtr uint32) (bool, error) {
	results, err := b.fnInstanceof.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(ctorPtr))
	if err != nil {
		return false, err
	}
	return int32(results[0]) > 0, nil
}

func (b *Bridge) Typeof(ctx context.Context, ctxPtr, valPtr uint32) (string, error) {
	results, err := b.fnTypeof.Call(ctx, uint64(ctxPtr), uint64(valPtr))
	if err != nil {
		return "", err
	}
	typeValPtr := uint32(results[0])
	return b.ToString(ctx, ctxPtr, typeValPtr)
}

// ============================================================================
// ArrayBuffer
// ============================================================================

func (b *Bridge) NewArrayBuffer(ctx context.Context, ctxPtr uint32, data []byte) (uint32, error) {
	var dataPtr uint32
	if len(data) > 0 {
		var err error
		dataPtr, err = b.WriteBytes(ctx, data)
		if err != nil {
			return 0, err
		}
	}
	results, err := b.fnNewArrayBuffer.Call(ctx, uint64(ctxPtr), uint64(dataPtr), uint64(len(data)))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) GetArrayBuffer(ctx context.Context, ctxPtr, valPtr uint32) ([]byte, error) {
	lenPtr, err := b.Alloc(ctx, 4)
	if err != nil {
		return nil, err
	}

	results, err := b.fnGetArrayBuffer.Call(ctx, uint64(ctxPtr), uint64(valPtr), uint64(lenPtr))
	if err != nil {
		return nil, err
	}
	bufPtr := uint32(results[0])
	if bufPtr == 0 {
		return nil, errors.New("not an ArrayBuffer")
	}

	lenBuf, ok := b.memory.Read(lenPtr, 4)
	if !ok {
		return nil, errors.New("failed to read length")
	}
	length := binary.LittleEndian.Uint32(lenBuf)

	return b.ReadBytes(bufPtr, length), nil
}

// ============================================================================
// C Function Binding (for Go callbacks)
// ============================================================================

// RegisterGoFunc registers a Go function and returns its function ID.
func (b *Bridge) RegisterGoFunc(fn GoFunc) uint32 {
	b.callbackMu.Lock()
	defer b.callbackMu.Unlock()

	funcID := b.nextFuncID
	b.nextFuncID++
	b.callbacks[funcID] = fn
	return funcID
}

// UnregisterGoFunc removes a registered Go function.
func (b *Bridge) UnregisterGoFunc(funcID uint32) {
	b.callbackMu.Lock()
	defer b.callbackMu.Unlock()
	delete(b.callbacks, funcID)
}

// NewCFunction creates a new JavaScript function that calls back to Go.
func (b *Bridge) NewCFunction(ctx context.Context, ctxPtr, funcID uint32, name string, argCount int32) (uint32, error) {
	namePtr, err := b.WriteString(ctx, name)
	if err != nil {
		return 0, err
	}
	results, err := b.fnNewCFunction.Call(ctx, uint64(ctxPtr), uint64(funcID), uint64(namePtr), uint64(argCount))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// ============================================================================
// Runtime Configuration
// ============================================================================

func (b *Bridge) SetMemoryLimit(ctx context.Context, rtPtr, limit uint32) error {
	_, err := b.fnSetMemoryLimit.Call(ctx, uint64(rtPtr), uint64(limit))
	return err
}

func (b *Bridge) SetMaxStackSize(ctx context.Context, rtPtr, stackSize uint32) error {
	_, err := b.fnSetMaxStackSize.Call(ctx, uint64(rtPtr), uint64(stackSize))
	return err
}

// ============================================================================
// Memory Info
// ============================================================================

func (b *Bridge) GetHeapPtr(ctx context.Context) (uint32, error) {
	results, err := b.fnGetHeapPtr.Call(ctx)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) GetHeapSize(ctx context.Context) (uint32, error) {
	results, err := b.fnGetHeapSize.Call(ctx)
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

func (b *Bridge) ResetHeap(ctx context.Context) error {
	_, err := b.fnResetHeap.Call(ctx)
	return err
}
