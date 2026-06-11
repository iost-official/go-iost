// Package quickjs provides Go bindings for the QuickJS-ng JavaScript engine.
//
// QuickJS-ng is a modern JavaScript engine with full ES2023+ support including
// classes, async/await, Promises, BigInt, and more. This package provides
// CGO-free bindings by compiling QuickJS-ng to WebAssembly and running it
// via wazero.
//
// Basic usage:
//
//	rt, err := quickjs.NewRuntime()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer rt.Close()
//
//	ctx, err := rt.NewContext()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer ctx.Close()
//
//	result, err := ctx.Eval("1 + 2")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.String()) // Output: 3
package quickjs

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/Gaurav-Gosain/quickjs/internal/bridge"
)

// EvalFlag represents flags for JavaScript evaluation.
type EvalFlag int32

const (
	// EvalGlobal evaluates code in global scope (default).
	EvalGlobal EvalFlag = 0
	// EvalModule evaluates code as an ES6 module.
	EvalModule EvalFlag = 1 << 0
)

// Runtime represents a JavaScript runtime instance.
// A runtime contains the WASM module and can create multiple contexts.
// All operations on a Runtime (and its Contexts) are serialized via a mutex
// because the underlying WASM execution is not thread-safe.
type Runtime struct {
	bridge  *bridge.Bridge
	rtPtr   uint32 // QuickJS runtime pointer
	goCtx   context.Context
	mu      sync.Mutex
	logFunc func(msg string)

	// For reentrant callback support: track which goroutine holds the lock
	lockHolder uintptr    // goroutine ID of current lock holder (0 if unlocked)
	lockDepth  int32      // recursion depth
	lockMu     sync.Mutex // protects lockHolder and lockDepth
}

// lock acquires the runtime mutex, supporting reentrant locking from callbacks.
func (r *Runtime) lock() {
	gid := getGoroutineID()

	r.lockMu.Lock()
	if r.lockHolder == gid {
		// Same goroutine - allow recursive lock
		r.lockDepth++
		r.lockMu.Unlock()
		return
	}
	r.lockMu.Unlock()

	// Different goroutine - need to acquire the actual mutex
	r.mu.Lock()

	r.lockMu.Lock()
	r.lockHolder = gid
	r.lockDepth = 1
	r.lockMu.Unlock()
}

// unlock releases the runtime mutex.
func (r *Runtime) unlock() {
	r.lockMu.Lock()
	r.lockDepth--
	if r.lockDepth == 0 {
		r.lockHolder = 0
		r.lockMu.Unlock()
		r.mu.Unlock()
	} else {
		r.lockMu.Unlock()
	}
}

// getGoroutineID returns a unique identifier for the current goroutine.
// This is a hack that reads from the runtime stack, but it's safe and fast.
func getGoroutineID() uintptr {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// Stack looks like "goroutine 123 [running]:\n..."
	// We parse the number after "goroutine "
	var id uintptr
	for i := 10; i < n && buf[i] != ' '; i++ {
		id = id*10 + uintptr(buf[i]-'0')
	}
	return id
}

// NewRuntime creates a new JavaScript runtime with default settings.
func NewRuntime() (*Runtime, error) {
	return NewRuntimeWithContext(context.Background())
}

// NewRuntimeWithContext creates a new JavaScript runtime with the given context.
func NewRuntimeWithContext(ctx context.Context) (*Runtime, error) {
	b, err := bridge.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize QuickJS bridge: %w", err)
	}

	// Create the QuickJS runtime
	rtPtr, err := b.NewRuntime(ctx)
	if err != nil {
		b.Close(ctx)
		return nil, fmt.Errorf("failed to create QuickJS runtime: %w", err)
	}

	return &Runtime{
		bridge:  b,
		rtPtr:   rtPtr,
		goCtx:   ctx,
		logFunc: func(msg string) { fmt.Print(msg) },
	}, nil
}

// Close releases all resources associated with the runtime.
func (r *Runtime) Close() error {
	r.lock()
	defer r.unlock()
	if err := r.bridge.FreeRuntime(r.goCtx, r.rtPtr); err != nil {
		return err
	}
	return r.bridge.Close(r.goCtx)
}

// SetLogFunc sets the function called for console.log output from JavaScript.
func (r *Runtime) SetLogFunc(fn func(msg string)) {
	r.lock()
	defer r.unlock()
	r.logFunc = fn
	r.bridge.SetLogFunc(fn)
}

// NewContext creates a new JavaScript execution context.
func (r *Runtime) NewContext() (*Context, error) {
	r.lock()
	defer r.unlock()

	ctxPtr, err := r.bridge.NewContext(r.goCtx, r.rtPtr)
	if err != nil {
		return nil, fmt.Errorf("failed to create JavaScript context: %w", err)
	}

	// Add console.log support
	if err := r.bridge.AddConsole(r.goCtx, ctxPtr); err != nil {
		_ = r.bridge.FreeContext(r.goCtx, ctxPtr)
		return nil, fmt.Errorf("failed to add console support: %w", err)
	}

	return &Context{
		runtime: r,
		ctxPtr:  ctxPtr,
	}, nil
}

// RunGC triggers garbage collection.
func (r *Runtime) RunGC() error {
	r.lock()
	defer r.unlock()
	return r.bridge.RunGC(r.goCtx, r.rtPtr)
}

// ExecutePendingJobs executes pending promise jobs.
// Returns the number of jobs executed, or an error.
func (r *Runtime) ExecutePendingJobs() (int, error) {
	r.lock()
	defer r.unlock()
	n, err := r.bridge.ExecutePendingJobs(r.goCtx, r.rtPtr)
	return int(n), err
}

// SetMemoryLimit sets the memory limit for the runtime in bytes.
func (r *Runtime) SetMemoryLimit(limit uint32) error {
	r.lock()
	defer r.unlock()
	return r.bridge.SetMemoryLimit(r.goCtx, r.rtPtr, limit)
}

// SetMaxStackSize sets the maximum stack size for the runtime.
func (r *Runtime) SetMaxStackSize(size uint32) error {
	r.lock()
	defer r.unlock()
	return r.bridge.SetMaxStackSize(r.goCtx, r.rtPtr, size)
}

// Context represents a JavaScript execution context.
type Context struct {
	runtime *Runtime
	ctxPtr  uint32
}

// Close releases all resources associated with the context.
func (c *Context) Close() error {
	c.runtime.lock()
	defer c.runtime.unlock()
	return c.runtime.bridge.FreeContext(c.runtime.goCtx, c.ctxPtr)
}

// Eval evaluates JavaScript code and returns the result.
func (c *Context) Eval(code string) (Value, error) {
	return c.EvalFile(code, "<eval>")
}

// EvalFile evaluates JavaScript code with a specified filename for error messages.
func (c *Context) EvalFile(code, filename string) (Value, error) {
	c.runtime.lock()
	defer c.runtime.unlock()

	valPtr, err := c.runtime.bridge.Eval(c.runtime.goCtx, c.ctxPtr, code, filename, int32(EvalGlobal))
	if err != nil {
		return Value{}, err
	}

	return c.checkException(valPtr)
}

// EvalModule evaluates JavaScript code as an ES6 module.
func (c *Context) EvalModule(code, filename string) (Value, error) {
	c.runtime.lock()
	defer c.runtime.unlock()

	valPtr, err := c.runtime.bridge.EvalModule(c.runtime.goCtx, c.ctxPtr, code, filename)
	if err != nil {
		return Value{}, err
	}

	return c.checkException(valPtr)
}

// checkException checks if the value is an exception and returns an error if so.
// Caller must hold the mutex.
func (c *Context) checkException(valPtr uint32) (Value, error) {
	isExc, _ := c.runtime.bridge.IsException(c.runtime.goCtx, valPtr)
	if isExc {
		// Get the actual exception
		excPtr, _ := c.runtime.bridge.GetException(c.runtime.goCtx, c.ctxPtr)
		errMsg, _ := c.runtime.bridge.GetErrorMessage(c.runtime.goCtx, c.ctxPtr, excPtr)
		if errMsg == "" {
			errMsg = "JavaScript exception"
		}
		_ = c.runtime.bridge.FreeValue(c.runtime.goCtx, c.ctxPtr, excPtr)
		return Value{}, errors.New(errMsg)
	}
	return Value{ctx: c, ptr: valPtr}, nil
}

// Global returns the global object.
func (c *Context) Global() (Value, error) {
	c.runtime.lock()
	defer c.runtime.unlock()

	valPtr, err := c.runtime.bridge.GetGlobalObject(c.runtime.goCtx, c.ctxPtr)
	if err != nil {
		return Value{}, err
	}
	return Value{ctx: c, ptr: valPtr}, nil
}

// ============================================================================
// Value Creation
// ============================================================================

// Undefined returns the JavaScript undefined value.
func (c *Context) Undefined() Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	return c.undefinedUnlocked()
}

// undefinedUnlocked returns undefined without acquiring lock (for callback use).
func (c *Context) undefinedUnlocked() Value {
	ptr, _ := c.runtime.bridge.NewUndefined(c.runtime.goCtx)
	return Value{ctx: c, ptr: ptr}
}

// Null returns the JavaScript null value.
func (c *Context) Null() Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewNull(c.runtime.goCtx)
	return Value{ctx: c, ptr: ptr}
}

// Bool creates a new JavaScript boolean.
func (c *Context) Bool(v bool) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewBool(c.runtime.goCtx, v)
	return Value{ctx: c, ptr: ptr}
}

// Int32 creates a new JavaScript integer from an int32.
func (c *Context) Int32(v int32) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewInt32(c.runtime.goCtx, v)
	return Value{ctx: c, ptr: ptr}
}

// Int64 creates a new JavaScript integer from an int64.
func (c *Context) Int64(v int64) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewInt64(c.runtime.goCtx, c.ctxPtr, v)
	return Value{ctx: c, ptr: ptr}
}

// Float64 creates a new JavaScript number from a float64.
func (c *Context) Float64(v float64) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewFloat64(c.runtime.goCtx, v)
	return Value{ctx: c, ptr: ptr}
}

// String creates a new JavaScript string.
func (c *Context) String(s string) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewString(c.runtime.goCtx, c.ctxPtr, s)
	return Value{ctx: c, ptr: ptr}
}

// StringLen creates a new JavaScript string, preserving null bytes.
func (c *Context) StringLen(s string) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewStringLen(c.runtime.goCtx, c.ctxPtr, s)
	return Value{ctx: c, ptr: ptr}
}

// Object creates a new JavaScript object.
func (c *Context) Object() Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewObject(c.runtime.goCtx, c.ctxPtr)
	return Value{ctx: c, ptr: ptr}
}

// Array creates a new JavaScript array.
func (c *Context) Array() Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewArray(c.runtime.goCtx, c.ctxPtr)
	return Value{ctx: c, ptr: ptr}
}

// BigInt creates a new JavaScript BigInt from an int64.
func (c *Context) BigInt(v int64) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewBigInt64(c.runtime.goCtx, c.ctxPtr, v)
	return Value{ctx: c, ptr: ptr}
}

// Date creates a new JavaScript Date from Unix milliseconds.
func (c *Context) Date(epochMs float64) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewDate(c.runtime.goCtx, c.ctxPtr, epochMs)
	return Value{ctx: c, ptr: ptr}
}

// ArrayBuffer creates a new JavaScript ArrayBuffer with the given data.
func (c *Context) ArrayBuffer(data []byte) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.NewArrayBuffer(c.runtime.goCtx, c.ctxPtr, data)
	return Value{ctx: c, ptr: ptr}
}

// ParseJSON parses a JSON string and returns the result.
func (c *Context) ParseJSON(json string) (Value, error) {
	c.runtime.lock()
	defer c.runtime.unlock()

	valPtr, err := c.runtime.bridge.JSONParse(c.runtime.goCtx, c.ctxPtr, json)
	if err != nil {
		return Value{}, err
	}
	return c.checkException(valPtr)
}

// ============================================================================
// Go Function Binding
// ============================================================================

// GoFunc is the signature for Go functions callable from JavaScript.
type GoFunc func(ctx *Context, this Value, args []Value) Value

// Function creates a new JavaScript function that calls the given Go function.
func (c *Context) Function(name string, fn GoFunc) Value {
	// Create wrapper that handles the bridge callback
	// Note: This callback runs while the mutex is already held by Eval,
	// so we must use unlocked methods here.
	bridgeFn := func(ctxPtr uint32, argPtrs []uint32) uint32 {
		args := make([]Value, len(argPtrs))
		for i, ptr := range argPtrs {
			args[i] = Value{ctx: c, ptr: ptr}
		}

		result := fn(c, c.undefinedUnlocked(), args)
		return result.ptr
	}

	// Register the callback
	funcID := c.runtime.bridge.RegisterGoFunc(bridgeFn)

	c.runtime.lock()
	defer c.runtime.unlock()

	ptr, err := c.runtime.bridge.NewCFunction(c.runtime.goCtx, c.ctxPtr, funcID, name, -1)
	if err != nil {
		c.runtime.bridge.UnregisterGoFunc(funcID)
		return c.undefinedUnlocked()
	}

	return Value{ctx: c, ptr: ptr}
}

// SetGlobal sets a value on the global object.
func (c *Context) SetGlobal(name string, val Value) error {
	c.runtime.lock()
	defer c.runtime.unlock()

	globalPtr, err := c.runtime.bridge.GetGlobalObject(c.runtime.goCtx, c.ctxPtr)
	if err != nil {
		return err
	}
	return c.runtime.bridge.SetProperty(c.runtime.goCtx, c.ctxPtr, globalPtr, name, val.ptr)
}

// GetGlobal gets a value from the global object.
func (c *Context) GetGlobal(name string) (Value, error) {
	c.runtime.lock()
	defer c.runtime.unlock()

	globalPtr, err := c.runtime.bridge.GetGlobalObject(c.runtime.goCtx, c.ctxPtr)
	if err != nil {
		return Value{}, err
	}
	valPtr, err := c.runtime.bridge.GetProperty(c.runtime.goCtx, c.ctxPtr, globalPtr, name)
	if err != nil {
		return Value{}, err
	}
	return Value{ctx: c, ptr: valPtr}, nil
}

// ThrowError throws a JavaScript error with the given message.
func (c *Context) ThrowError(msg string) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.ThrowError(c.runtime.goCtx, c.ctxPtr, msg)
	return Value{ctx: c, ptr: ptr}
}

// ThrowTypeError throws a JavaScript TypeError with the given message.
func (c *Context) ThrowTypeError(msg string) Value {
	c.runtime.lock()
	defer c.runtime.unlock()
	ptr, _ := c.runtime.bridge.ThrowTypeError(c.runtime.goCtx, c.ctxPtr, msg)
	return Value{ctx: c, ptr: ptr}
}

// ============================================================================
// Value
// ============================================================================

// Value represents a JavaScript value.
type Value struct {
	ctx *Context
	ptr uint32
}

// IsUndefined returns true if the value is undefined.
func (v Value) IsUndefined() bool {
	if v.ctx == nil {
		return true
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsUndefined(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsNull returns true if the value is null.
func (v Value) IsNull() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsNull(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsBool returns true if the value is a boolean.
func (v Value) IsBool() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsBool(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsNumber returns true if the value is a number.
func (v Value) IsNumber() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsNumber(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsString returns true if the value is a string.
func (v Value) IsString() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsString(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsSymbol returns true if the value is a symbol.
func (v Value) IsSymbol() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsSymbol(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsObject returns true if the value is an object.
func (v Value) IsObject() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsObject(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsArray returns true if the value is an array.
func (v Value) IsArray() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsArray(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsFunction returns true if the value is a function.
func (v Value) IsFunction() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsFunction(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return result
}

// IsError returns true if the value is an Error object.
func (v Value) IsError() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsError(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsBigInt returns true if the value is a BigInt.
func (v Value) IsBigInt() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsBigInt(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsDate returns true if the value is a Date.
func (v Value) IsDate() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsDate(v.ctx.runtime.goCtx, v.ptr)
	return result
}

// IsPromise returns true if the value is a Promise (has a 'then' method).
func (v Value) IsPromise() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.IsPromise(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return result
}

// ============================================================================
// Value Conversion
// ============================================================================

// String returns the string representation of the value.
func (v Value) String() string {
	if v.ctx == nil {
		return "undefined"
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	s, _ := v.ctx.runtime.bridge.ToString(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return s
}

// StringBytes returns the string representation of the value, preserving null bytes.
func (v Value) StringBytes() string {
	if v.ctx == nil {
		return "undefined"
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	s, _ := v.ctx.runtime.bridge.ToStringLen(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return s
}

// Bool returns the value as a boolean.
func (v Value) Bool() bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	b, _ := v.ctx.runtime.bridge.ToBool(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return b
}

// Int32 returns the value as an int32.
func (v Value) Int32() (int32, error) {
	if v.ctx == nil {
		return 0, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.ToInt32(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// Int64 returns the value as an int64.
func (v Value) Int64() (int64, error) {
	if v.ctx == nil {
		return 0, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.ToInt64(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// Float64 returns the value as a float64.
func (v Value) Float64() (float64, error) {
	if v.ctx == nil {
		return 0, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.ToFloat64(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// BigInt returns the value as an int64 (for BigInt values).
func (v Value) BigInt() (int64, error) {
	if v.ctx == nil {
		return 0, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.ToBigInt64(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// JSONStringify returns the JSON representation of the value.
func (v Value) JSONStringify() (string, error) {
	if v.ctx == nil {
		return "", errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.JSONStringify(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// Bytes returns the value as bytes (for ArrayBuffer values).
func (v Value) Bytes() ([]byte, error) {
	if v.ctx == nil {
		return nil, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.GetArrayBuffer(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
}

// Typeof returns the JavaScript typeof string for the value.
func (v Value) Typeof() string {
	if v.ctx == nil {
		return "undefined"
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	s, _ := v.ctx.runtime.bridge.Typeof(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr)
	return s
}

// ============================================================================
// Object Operations
// ============================================================================

// Get returns a property value by name.
func (v Value) Get(prop string) (Value, error) {
	if v.ctx == nil {
		return Value{}, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	valPtr, err := v.ctx.runtime.bridge.GetProperty(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, prop)
	if err != nil {
		return Value{}, err
	}
	return Value{ctx: v.ctx, ptr: valPtr}, nil
}

// Set sets a property value by name.
func (v Value) Set(prop string, val Value) error {
	if v.ctx == nil {
		return errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.SetProperty(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, prop, val.ptr)
}

// Has returns true if the object has the given property.
func (v Value) Has(prop string) bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.HasProperty(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, prop)
	return result
}

// Delete deletes a property by name.
func (v Value) Delete(prop string) error {
	if v.ctx == nil {
		return errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.DeleteProperty(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, prop)
}

// GetIdx returns an element by index (for arrays).
func (v Value) GetIdx(idx int) (Value, error) {
	if v.ctx == nil {
		return Value{}, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	valPtr, err := v.ctx.runtime.bridge.GetPropertyUint32(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, uint32(idx))
	if err != nil {
		return Value{}, err
	}
	return Value{ctx: v.ctx, ptr: valPtr}, nil
}

// SetIdx sets an element by index (for arrays).
func (v Value) SetIdx(idx int, val Value) error {
	if v.ctx == nil {
		return errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	return v.ctx.runtime.bridge.SetPropertyUint32(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, uint32(idx), val.ptr)
}

// Len returns the length property of the value (for arrays/strings).
func (v Value) Len() int {
	if v.ctx == nil {
		return 0
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	lenPtr, err := v.ctx.runtime.bridge.GetProperty(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, "length")
	if err != nil {
		return 0
	}
	n, _ := v.ctx.runtime.bridge.ToInt32(v.ctx.runtime.goCtx, v.ctx.ctxPtr, lenPtr)
	return int(n)
}

// ============================================================================
// Function Calling
// ============================================================================

// Call calls the value as a function with the given arguments.
func (v Value) Call(this Value, args ...Value) (Value, error) {
	if v.ctx == nil {
		return Value{}, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()

	argPtrs := make([]uint32, len(args))
	for i, arg := range args {
		argPtrs[i] = arg.ptr
	}

	resultPtr, err := v.ctx.runtime.bridge.Call(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, this.ptr, argPtrs)
	if err != nil {
		return Value{}, err
	}

	return v.ctx.checkException(resultPtr)
}

// CallMethod calls a method on the value with the given arguments.
func (v Value) CallMethod(method string, args ...Value) (Value, error) {
	if v.ctx == nil {
		return Value{}, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()

	argPtrs := make([]uint32, len(args))
	for i, arg := range args {
		argPtrs[i] = arg.ptr
	}

	resultPtr, err := v.ctx.runtime.bridge.Invoke(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, method, argPtrs)
	if err != nil {
		return Value{}, err
	}

	return v.ctx.checkException(resultPtr)
}

// New calls the value as a constructor with the given arguments.
func (v Value) New(args ...Value) (Value, error) {
	if v.ctx == nil {
		return Value{}, errors.New("nil value")
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()

	argPtrs := make([]uint32, len(args))
	for i, arg := range args {
		argPtrs[i] = arg.ptr
	}

	resultPtr, err := v.ctx.runtime.bridge.CallConstructor(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, argPtrs)
	if err != nil {
		return Value{}, err
	}

	return v.ctx.checkException(resultPtr)
}

// Instanceof returns true if the value is an instance of the given constructor.
func (v Value) Instanceof(ctor Value) bool {
	if v.ctx == nil {
		return false
	}
	v.ctx.runtime.lock()
	defer v.ctx.runtime.unlock()
	result, _ := v.ctx.runtime.bridge.Instanceof(v.ctx.runtime.goCtx, v.ctx.ctxPtr, v.ptr, ctor.ptr)
	return result
}
