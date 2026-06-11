/**
 * QuickJS-ng WASM Bridge
 * 
 * This bridge provides a clean interface between Go (via wazero) and QuickJS-ng.
 * 
 * Key design decisions:
 * - JSValue is 64-bit (NAN boxing) but WASM is 32-bit, so we store JSValues
 *   in memory and pass pointers (uint32) instead of raw values.
 * - We use a slot-based allocator with freelist for JSValue storage.
 * - Temporary allocations (strings from Go) use a simple arena that can be reset.
 * - QuickJS-ng uses the default libc malloc.
 */

#include "quickjs-ng/quickjs.h"
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

// ============================================================================
// JSValue Slot Storage (with freelist for reuse)
// ============================================================================

#define MAX_JSVALUE_SLOTS 65536  // 64K slots = 512KB for JSValues

typedef struct {
    JSValue value;
    uint32_t next_free;  // Index of next free slot (0 = end of list)
} JSValueSlot;

static JSValueSlot jsvalue_slots[MAX_JSVALUE_SLOTS];
static uint32_t first_free_slot = 0;  // Head of freelist
static int slots_initialized = 0;

static void init_jsvalue_slots(void) {
    if (slots_initialized) return;
    // Initialize freelist: each slot points to the next
    for (uint32_t i = 0; i < MAX_JSVALUE_SLOTS - 1; i++) {
        jsvalue_slots[i].next_free = i + 1;
        jsvalue_slots[i].value = JS_UNDEFINED;
    }
    jsvalue_slots[MAX_JSVALUE_SLOTS - 1].next_free = 0;  // End of list
    first_free_slot = 1;  // Slot 0 is reserved (represents NULL)
    slots_initialized = 1;
}

// Store a JSValue and return slot index (1-based, 0 = NULL/error)
static uint32_t store_jsvalue(JSValue val) {
    init_jsvalue_slots();
    
    if (first_free_slot == 0) {
        // No free slots
        return 0;
    }
    
    uint32_t slot = first_free_slot;
    first_free_slot = jsvalue_slots[slot].next_free;
    jsvalue_slots[slot].value = val;
    jsvalue_slots[slot].next_free = 0;  // Mark as in-use
    
    return slot;
}

// Load a JSValue from slot index
static JSValue load_jsvalue(uint32_t slot) {
    if (slot == 0 || slot >= MAX_JSVALUE_SLOTS) return JS_UNDEFINED;
    return jsvalue_slots[slot].value;
}

// Free a JSValue slot (return to freelist)
static void free_jsvalue_slot(uint32_t slot) {
    if (slot == 0 || slot >= MAX_JSVALUE_SLOTS) return;
    jsvalue_slots[slot].value = JS_UNDEFINED;
    jsvalue_slots[slot].next_free = first_free_slot;
    first_free_slot = slot;
}

// ============================================================================
// Temporary Arena for Go string allocations
// ============================================================================

#define ARENA_SIZE (4 * 1024 * 1024)  // 4MB arena for temp strings
static char arena[ARENA_SIZE];
static size_t arena_ptr = 0;

// Allocate from arena (for temporary data from Go)
static void* arena_alloc(size_t size) {
    size = (size + 7) & ~7;  // 8-byte alignment
    if (arena_ptr + size > ARENA_SIZE) {
        // Arena full, reset it (assumes previous allocations are no longer needed)
        arena_ptr = 0;
    }
    void* ptr = &arena[arena_ptr];
    arena_ptr += size;
    return ptr;
}

// ============================================================================
// Host Function Import (for console.log, etc.)
// ============================================================================

// Imported from host (Go)
__attribute__((import_module("env"), import_name("host_log")))
extern void host_log(uint32_t ptr, uint32_t len);

// Host function for calling Go callbacks
__attribute__((import_module("env"), import_name("host_call_go")))
extern uint32_t host_call_go(uint32_t ctx_ptr, uint32_t func_id, int32_t argc, uint32_t argv_ptr);

// ============================================================================
// Runtime and Context Management
// ============================================================================

__attribute__((export_name("qjs_new_runtime")))
uint32_t qjs_new_runtime(void) {
    init_jsvalue_slots();
    JSRuntime* rt = JS_NewRuntime();
    if (!rt) return 0;
    return (uint32_t)(uintptr_t)rt;
}

__attribute__((export_name("qjs_free_runtime")))
void qjs_free_runtime(uint32_t rt_ptr) {
    if (!rt_ptr) return;
    JS_FreeRuntime((JSRuntime*)(uintptr_t)rt_ptr);
}

__attribute__((export_name("qjs_new_context")))
uint32_t qjs_new_context(uint32_t rt_ptr) {
    if (!rt_ptr) return 0;
    JSRuntime* rt = (JSRuntime*)(uintptr_t)rt_ptr;
    JSContext* ctx = JS_NewContext(rt);
    if (!ctx) return 0;
    return (uint32_t)(uintptr_t)ctx;
}

__attribute__((export_name("qjs_free_context")))
void qjs_free_context(uint32_t ctx_ptr) {
    if (!ctx_ptr) return;
    JS_FreeContext((JSContext*)(uintptr_t)ctx_ptr);
}

__attribute__((export_name("qjs_get_runtime")))
uint32_t qjs_get_runtime(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return (uint32_t)(uintptr_t)JS_GetRuntime(ctx);
}

// ============================================================================
// Memory Allocation (for Go to write strings/data into WASM memory)
// ============================================================================

__attribute__((export_name("qjs_alloc")))
uint32_t qjs_alloc(uint32_t size) {
    return (uint32_t)(uintptr_t)arena_alloc(size);
}

__attribute__((export_name("qjs_free")))
void qjs_free(uint32_t ptr) {
    // Arena allocator doesn't free individual allocations
    (void)ptr;
}

__attribute__((export_name("qjs_get_heap_ptr")))
uint32_t qjs_get_heap_ptr(void) {
    return (uint32_t)arena_ptr;
}

__attribute__((export_name("qjs_get_heap_size")))
uint32_t qjs_get_heap_size(void) {
    return ARENA_SIZE;
}

__attribute__((export_name("qjs_reset_heap")))
void qjs_reset_heap(void) {
    arena_ptr = 0;
}

// ============================================================================
// Evaluation
// ============================================================================

__attribute__((export_name("qjs_eval")))
uint32_t qjs_eval(uint32_t ctx_ptr, uint32_t code_ptr, uint32_t code_len, 
                  uint32_t filename_ptr, int32_t flags) {
    if (!ctx_ptr || !code_ptr) return 0;
    
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* code = (const char*)(uintptr_t)code_ptr;
    const char* filename = filename_ptr ? (const char*)(uintptr_t)filename_ptr : "<eval>";
    
    JSValue result = JS_Eval(ctx, code, code_len, filename, flags);
    return store_jsvalue(result);
}

__attribute__((export_name("qjs_eval_module")))
uint32_t qjs_eval_module(uint32_t ctx_ptr, uint32_t code_ptr, uint32_t code_len,
                         uint32_t filename_ptr) {
    if (!ctx_ptr || !code_ptr) return 0;
    
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* code = (const char*)(uintptr_t)code_ptr;
    const char* filename = filename_ptr ? (const char*)(uintptr_t)filename_ptr : "<module>";
    
    JSValue result = JS_Eval(ctx, code, code_len, filename, JS_EVAL_TYPE_MODULE);
    return store_jsvalue(result);
}

// ============================================================================
// Type Checking
// ============================================================================

__attribute__((export_name("qjs_is_exception")))
int32_t qjs_is_exception(uint32_t val_ptr) {
    return JS_IsException(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_undefined")))
int32_t qjs_is_undefined(uint32_t val_ptr) {
    return JS_IsUndefined(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_null")))
int32_t qjs_is_null(uint32_t val_ptr) {
    return JS_IsNull(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_bool")))
int32_t qjs_is_bool(uint32_t val_ptr) {
    return JS_IsBool(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_number")))
int32_t qjs_is_number(uint32_t val_ptr) {
    return JS_IsNumber(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_string")))
int32_t qjs_is_string(uint32_t val_ptr) {
    return JS_IsString(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_symbol")))
int32_t qjs_is_symbol(uint32_t val_ptr) {
    return JS_IsSymbol(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_object")))
int32_t qjs_is_object(uint32_t val_ptr) {
    return JS_IsObject(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_function")))
int32_t qjs_is_function(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return JS_IsFunction(ctx, load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_array")))
int32_t qjs_is_array(uint32_t val_ptr) {
    return JS_IsArray(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_error")))
int32_t qjs_is_error(uint32_t val_ptr) {
    return JS_IsError(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_big_int")))
int32_t qjs_is_big_int(uint32_t val_ptr) {
    return JS_IsBigInt(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_date")))
int32_t qjs_is_date(uint32_t val_ptr) {
    return JS_IsDate(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_regexp")))
int32_t qjs_is_regexp(uint32_t val_ptr) {
    return JS_IsRegExp(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_map")))
int32_t qjs_is_map(uint32_t val_ptr) {
    return JS_IsMap(load_jsvalue(val_ptr)) ? 1 : 0;
}

__attribute__((export_name("qjs_is_set")))
int32_t qjs_is_set(uint32_t val_ptr) {
    return JS_IsSet(load_jsvalue(val_ptr)) ? 1 : 0;
}

// ============================================================================
// Value Conversion - Getting values from JS
// ============================================================================

__attribute__((export_name("qjs_to_bool")))
int32_t qjs_to_bool(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return JS_ToBool(ctx, load_jsvalue(val_ptr));
}

__attribute__((export_name("qjs_to_int32")))
int32_t qjs_to_int32(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t result_ptr) {
    if (!ctx_ptr || !result_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    int32_t* result = (int32_t*)(uintptr_t)result_ptr;
    return JS_ToInt32(ctx, result, load_jsvalue(val_ptr));
}

__attribute__((export_name("qjs_to_int64")))
int32_t qjs_to_int64(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t result_ptr) {
    if (!ctx_ptr || !result_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    int64_t* result = (int64_t*)(uintptr_t)result_ptr;
    return JS_ToInt64(ctx, result, load_jsvalue(val_ptr));
}

__attribute__((export_name("qjs_to_float64")))
int32_t qjs_to_float64(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t result_ptr) {
    if (!ctx_ptr || !result_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    double* result = (double*)(uintptr_t)result_ptr;
    return JS_ToFloat64(ctx, result, load_jsvalue(val_ptr));
}

__attribute__((export_name("qjs_to_cstring")))
uint32_t qjs_to_cstring(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* str = JS_ToCString(ctx, load_jsvalue(val_ptr));
    return (uint32_t)(uintptr_t)str;
}

__attribute__((export_name("qjs_free_cstring")))
void qjs_free_cstring(uint32_t ctx_ptr, uint32_t str_ptr) {
    if (!ctx_ptr || !str_ptr) return;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JS_FreeCString(ctx, (const char*)(uintptr_t)str_ptr);
}

// Get string with length (for binary-safe strings)
__attribute__((export_name("qjs_to_cstring_len")))
uint32_t qjs_to_cstring_len(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t len_ptr) {
    if (!ctx_ptr || !len_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    size_t* len = (size_t*)(uintptr_t)len_ptr;
    const char* str = JS_ToCStringLen(ctx, len, load_jsvalue(val_ptr));
    return (uint32_t)(uintptr_t)str;
}

// ============================================================================
// Value Creation - Creating JS values from native types
// ============================================================================

__attribute__((export_name("qjs_new_undefined")))
uint32_t qjs_new_undefined(void) {
    return store_jsvalue(JS_UNDEFINED);
}

__attribute__((export_name("qjs_new_null")))
uint32_t qjs_new_null(void) {
    return store_jsvalue(JS_NULL);
}

__attribute__((export_name("qjs_new_bool")))
uint32_t qjs_new_bool(int32_t val) {
    return store_jsvalue(JS_NewBool(NULL, val));
}

__attribute__((export_name("qjs_new_int32")))
uint32_t qjs_new_int32(int32_t val) {
    return store_jsvalue(JS_NewInt32(NULL, val));
}

__attribute__((export_name("qjs_new_int64")))
uint32_t qjs_new_int64(uint32_t ctx_ptr, int64_t val) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewInt64(ctx, val));
}

__attribute__((export_name("qjs_new_float64")))
uint32_t qjs_new_float64(double val) {
    return store_jsvalue(JS_NewFloat64(NULL, val));
}

__attribute__((export_name("qjs_new_string")))
uint32_t qjs_new_string(uint32_t ctx_ptr, uint32_t str_ptr) {
    if (!ctx_ptr || !str_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* str = (const char*)(uintptr_t)str_ptr;
    return store_jsvalue(JS_NewString(ctx, str));
}

__attribute__((export_name("qjs_new_string_len")))
uint32_t qjs_new_string_len(uint32_t ctx_ptr, uint32_t str_ptr, uint32_t len) {
    if (!ctx_ptr || !str_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* str = (const char*)(uintptr_t)str_ptr;
    return store_jsvalue(JS_NewStringLen(ctx, str, len));
}

// ============================================================================
// Object Operations
// ============================================================================

__attribute__((export_name("qjs_new_object")))
uint32_t qjs_new_object(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewObject(ctx));
}

__attribute__((export_name("qjs_new_array")))
uint32_t qjs_new_array(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewArray(ctx));
}

__attribute__((export_name("qjs_get_property")))
uint32_t qjs_get_property(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t prop_ptr) {
    if (!ctx_ptr || !prop_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    const char* prop = (const char*)(uintptr_t)prop_ptr;
    return store_jsvalue(JS_GetPropertyStr(ctx, obj, prop));
}

__attribute__((export_name("qjs_set_property")))
int32_t qjs_set_property(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t prop_ptr, uint32_t val_ptr) {
    if (!ctx_ptr || !prop_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    const char* prop = (const char*)(uintptr_t)prop_ptr;
    JSValue val = load_jsvalue(val_ptr);
    return JS_SetPropertyStr(ctx, obj, prop, JS_DupValue(ctx, val));
}

__attribute__((export_name("qjs_has_property")))
int32_t qjs_has_property(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t prop_ptr) {
    if (!ctx_ptr || !prop_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    const char* prop = (const char*)(uintptr_t)prop_ptr;
    JSAtom atom = JS_NewAtom(ctx, prop);
    int result = JS_HasProperty(ctx, obj, atom);
    JS_FreeAtom(ctx, atom);
    return result;
}

__attribute__((export_name("qjs_delete_property")))
int32_t qjs_delete_property(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t prop_ptr) {
    if (!ctx_ptr || !prop_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    const char* prop = (const char*)(uintptr_t)prop_ptr;
    JSAtom atom = JS_NewAtom(ctx, prop);
    int result = JS_DeleteProperty(ctx, obj, atom, 0);
    JS_FreeAtom(ctx, atom);
    return result;
}

// Index-based property access (for arrays)
__attribute__((export_name("qjs_get_property_uint32")))
uint32_t qjs_get_property_uint32(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t idx) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    return store_jsvalue(JS_GetPropertyUint32(ctx, obj, idx));
}

__attribute__((export_name("qjs_set_property_uint32")))
int32_t qjs_set_property_uint32(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t idx, uint32_t val_ptr) {
    if (!ctx_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    JSValue val = load_jsvalue(val_ptr);
    return JS_SetPropertyUint32(ctx, obj, idx, JS_DupValue(ctx, val));
}

// ============================================================================
// Global Object
// ============================================================================

__attribute__((export_name("qjs_get_global_object")))
uint32_t qjs_get_global_object(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_GetGlobalObject(ctx));
}

// ============================================================================
// Function Calling
// ============================================================================

__attribute__((export_name("qjs_call")))
uint32_t qjs_call(uint32_t ctx_ptr, uint32_t func_ptr, uint32_t this_ptr, 
                  int32_t argc, uint32_t argv_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue func = load_jsvalue(func_ptr);
    JSValue this_val = load_jsvalue(this_ptr);
    
    // Load arguments from memory
    JSValue* argv = NULL;
    if (argc > 0 && argv_ptr) {
        argv = (JSValue*)arena_alloc(sizeof(JSValue) * argc);
        if (!argv) return store_jsvalue(JS_EXCEPTION);
        
        uint32_t* arg_ptrs = (uint32_t*)(uintptr_t)argv_ptr;
        for (int i = 0; i < argc; i++) {
            argv[i] = load_jsvalue(arg_ptrs[i]);
        }
    }
    
    JSValue result = JS_Call(ctx, func, this_val, argc, argv);
    return store_jsvalue(result);
}

__attribute__((export_name("qjs_call_constructor")))
uint32_t qjs_call_constructor(uint32_t ctx_ptr, uint32_t func_ptr, 
                               int32_t argc, uint32_t argv_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue func = load_jsvalue(func_ptr);
    
    // Load arguments from memory
    JSValue* argv = NULL;
    if (argc > 0 && argv_ptr) {
        argv = (JSValue*)arena_alloc(sizeof(JSValue) * argc);
        if (!argv) return store_jsvalue(JS_EXCEPTION);
        
        uint32_t* arg_ptrs = (uint32_t*)(uintptr_t)argv_ptr;
        for (int i = 0; i < argc; i++) {
            argv[i] = load_jsvalue(arg_ptrs[i]);
        }
    }
    
    JSValue result = JS_CallConstructor(ctx, func, argc, argv);
    return store_jsvalue(result);
}

__attribute__((export_name("qjs_invoke")))
uint32_t qjs_invoke(uint32_t ctx_ptr, uint32_t obj_ptr, uint32_t method_ptr,
                    int32_t argc, uint32_t argv_ptr) {
    if (!ctx_ptr || !method_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    const char* method = (const char*)(uintptr_t)method_ptr;
    
    // Load arguments from memory
    JSValue* argv = NULL;
    if (argc > 0 && argv_ptr) {
        argv = (JSValue*)arena_alloc(sizeof(JSValue) * argc);
        if (!argv) return store_jsvalue(JS_EXCEPTION);
        
        uint32_t* arg_ptrs = (uint32_t*)(uintptr_t)argv_ptr;
        for (int i = 0; i < argc; i++) {
            argv[i] = load_jsvalue(arg_ptrs[i]);
        }
    }
    
    JSAtom atom = JS_NewAtom(ctx, method);
    JSValue result = JS_Invoke(ctx, obj, atom, argc, argv);
    JS_FreeAtom(ctx, atom);
    return store_jsvalue(result);
}

// ============================================================================
// Exception Handling
// ============================================================================

__attribute__((export_name("qjs_get_exception")))
uint32_t qjs_get_exception(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_GetException(ctx));
}

__attribute__((export_name("qjs_has_exception")))
int32_t qjs_has_exception(uint32_t ctx_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return JS_HasException(ctx) ? 1 : 0;
}

__attribute__((export_name("qjs_throw")))
uint32_t qjs_throw(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    return store_jsvalue(JS_Throw(ctx, JS_DupValue(ctx, val)));
}

__attribute__((export_name("qjs_throw_error")))
uint32_t qjs_throw_error(uint32_t ctx_ptr, uint32_t msg_ptr) {
    if (!ctx_ptr || !msg_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* msg = (const char*)(uintptr_t)msg_ptr;
    return store_jsvalue(JS_ThrowInternalError(ctx, "%s", msg));
}

__attribute__((export_name("qjs_throw_type_error")))
uint32_t qjs_throw_type_error(uint32_t ctx_ptr, uint32_t msg_ptr) {
    if (!ctx_ptr || !msg_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* msg = (const char*)(uintptr_t)msg_ptr;
    return store_jsvalue(JS_ThrowTypeError(ctx, "%s", msg));
}

__attribute__((export_name("qjs_throw_range_error")))
uint32_t qjs_throw_range_error(uint32_t ctx_ptr, uint32_t msg_ptr) {
    if (!ctx_ptr || !msg_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* msg = (const char*)(uintptr_t)msg_ptr;
    return store_jsvalue(JS_ThrowRangeError(ctx, "%s", msg));
}

__attribute__((export_name("qjs_throw_syntax_error")))
uint32_t qjs_throw_syntax_error(uint32_t ctx_ptr, uint32_t msg_ptr) {
    if (!ctx_ptr || !msg_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* msg = (const char*)(uintptr_t)msg_ptr;
    return store_jsvalue(JS_ThrowSyntaxError(ctx, "%s", msg));
}

__attribute__((export_name("qjs_throw_reference_error")))
uint32_t qjs_throw_reference_error(uint32_t ctx_ptr, uint32_t msg_ptr) {
    if (!ctx_ptr || !msg_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* msg = (const char*)(uintptr_t)msg_ptr;
    return store_jsvalue(JS_ThrowReferenceError(ctx, "%s", msg));
}

// ============================================================================
// Value Management
// ============================================================================

__attribute__((export_name("qjs_dup_value")))
uint32_t qjs_dup_value(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    return store_jsvalue(JS_DupValue(ctx, val));
}

__attribute__((export_name("qjs_free_value")))
void qjs_free_value(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr || !val_ptr) return;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    JS_FreeValue(ctx, val);
    // Return the slot to the freelist
    free_jsvalue_slot(val_ptr);
}

// ============================================================================
// JSON
// ============================================================================

__attribute__((export_name("qjs_json_parse")))
uint32_t qjs_json_parse(uint32_t ctx_ptr, uint32_t json_ptr, uint32_t len) {
    if (!ctx_ptr || !json_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* json = (const char*)(uintptr_t)json_ptr;
    return store_jsvalue(JS_ParseJSON(ctx, json, len, "<json>"));
}

__attribute__((export_name("qjs_json_stringify")))
uint32_t qjs_json_stringify(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    return store_jsvalue(JS_JSONStringify(ctx, val, JS_UNDEFINED, JS_UNDEFINED));
}

// ============================================================================
// Garbage Collection
// ============================================================================

__attribute__((export_name("qjs_run_gc")))
void qjs_run_gc(uint32_t rt_ptr) {
    if (!rt_ptr) return;
    JSRuntime* rt = (JSRuntime*)(uintptr_t)rt_ptr;
    JS_RunGC(rt);
}

// ============================================================================
// Promise Handling
// ============================================================================

__attribute__((export_name("qjs_is_promise")))
int32_t qjs_is_promise(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    
    // Check if it's an object with a 'then' method that is a function
    if (!JS_IsObject(val)) return 0;
    
    JSValue then_val = JS_GetPropertyStr(ctx, val, "then");
    int is_promise = JS_IsFunction(ctx, then_val);
    JS_FreeValue(ctx, then_val);
    return is_promise;
}

__attribute__((export_name("qjs_new_promise")))
uint32_t qjs_new_promise(uint32_t ctx_ptr, uint32_t resolving_funcs_ptr) {
    if (!ctx_ptr || !resolving_funcs_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    
    JSValue funcs[2];
    JSValue promise = JS_NewPromiseCapability(ctx, funcs);
    
    // Store resolve and reject functions
    uint32_t* out = (uint32_t*)(uintptr_t)resolving_funcs_ptr;
    out[0] = store_jsvalue(funcs[0]);
    out[1] = store_jsvalue(funcs[1]);
    
    return store_jsvalue(promise);
}

__attribute__((export_name("qjs_execute_pending_jobs")))
int32_t qjs_execute_pending_jobs(uint32_t rt_ptr) {
    if (!rt_ptr) return -1;
    JSRuntime* rt = (JSRuntime*)(uintptr_t)rt_ptr;
    JSContext* pctx;
    int ret;
    
    // Execute all pending jobs
    while ((ret = JS_ExecutePendingJob(rt, &pctx)) > 0) {
        // Job executed
    }
    
    return ret;
}

// ============================================================================
// BigInt Support
// ============================================================================

__attribute__((export_name("qjs_new_big_int64")))
uint32_t qjs_new_big_int64(uint32_t ctx_ptr, int64_t val) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewBigInt64(ctx, val));
}

__attribute__((export_name("qjs_new_big_uint64")))
uint32_t qjs_new_big_uint64(uint32_t ctx_ptr, uint64_t val) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewBigUint64(ctx, val));
}

__attribute__((export_name("qjs_to_big_int64")))
int32_t qjs_to_big_int64(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t result_ptr) {
    if (!ctx_ptr || !result_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    int64_t* result = (int64_t*)(uintptr_t)result_ptr;
    return JS_ToBigInt64(ctx, result, load_jsvalue(val_ptr));
}

// ============================================================================
// Date Support
// ============================================================================

__attribute__((export_name("qjs_new_date")))
uint32_t qjs_new_date(uint32_t ctx_ptr, double epoch_ms) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_NewDate(ctx, epoch_ms));
}

// ============================================================================
// Instanceof and typeof
// ============================================================================

__attribute__((export_name("qjs_instanceof")))
int32_t qjs_instanceof(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t ctor_ptr) {
    if (!ctx_ptr) return -1;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return JS_IsInstanceOf(ctx, load_jsvalue(val_ptr), load_jsvalue(ctor_ptr));
}

__attribute__((export_name("qjs_typeof")))
uint32_t qjs_typeof(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue val = load_jsvalue(val_ptr);
    
    const char* type_str;
    if (JS_IsUndefined(val)) {
        type_str = "undefined";
    } else if (JS_IsNull(val)) {
        type_str = "object";  // typeof null === "object" in JS
    } else if (JS_IsBool(val)) {
        type_str = "boolean";
    } else if (JS_IsNumber(val)) {
        type_str = "number";
    } else if (JS_IsString(val)) {
        type_str = "string";
    } else if (JS_IsSymbol(val)) {
        type_str = "symbol";
    } else if (JS_IsBigInt(val)) {
        type_str = "bigint";
    } else if (JS_IsFunction(ctx, val)) {
        type_str = "function";
    } else {
        type_str = "object";
    }
    
    return store_jsvalue(JS_NewString(ctx, type_str));
}

// ============================================================================
// Object Property Enumeration
// ============================================================================

__attribute__((export_name("qjs_get_own_property_names")))
uint32_t qjs_get_own_property_names(uint32_t ctx_ptr, uint32_t obj_ptr, 
                                     uint32_t count_ptr, int32_t flags) {
    if (!ctx_ptr || !count_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue obj = load_jsvalue(obj_ptr);
    
    JSPropertyEnum* props;
    uint32_t prop_count;
    
    if (JS_GetOwnPropertyNames(ctx, &props, &prop_count, obj, flags) < 0) {
        return 0;
    }
    
    // Store count
    *(uint32_t*)(uintptr_t)count_ptr = prop_count;
    
    // Create array of property names
    JSValue arr = JS_NewArray(ctx);
    for (uint32_t i = 0; i < prop_count; i++) {
        JSValue name = JS_AtomToString(ctx, props[i].atom);
        JS_SetPropertyUint32(ctx, arr, i, name);
        JS_FreeAtom(ctx, props[i].atom);
    }
    js_free(ctx, props);
    
    return store_jsvalue(arr);
}

// ============================================================================
// ArrayBuffer Support
// ============================================================================

__attribute__((export_name("qjs_new_array_buffer")))
uint32_t qjs_new_array_buffer(uint32_t ctx_ptr, uint32_t data_ptr, uint32_t len) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    
    uint8_t* data = data_ptr ? (uint8_t*)(uintptr_t)data_ptr : NULL;
    return store_jsvalue(JS_NewArrayBufferCopy(ctx, data, len));
}

__attribute__((export_name("qjs_get_array_buffer")))
uint32_t qjs_get_array_buffer(uint32_t ctx_ptr, uint32_t val_ptr, uint32_t len_ptr) {
    if (!ctx_ptr || !len_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    size_t* len = (size_t*)(uintptr_t)len_ptr;
    
    uint8_t* buf = JS_GetArrayBuffer(ctx, len, load_jsvalue(val_ptr));
    return (uint32_t)(uintptr_t)buf;
}

// ============================================================================
// Console/Print Support (using host_log)
// ============================================================================

static JSValue js_print(JSContext *ctx, JSValue this_val, int argc, JSValue *argv) {
    for (int i = 0; i < argc; i++) {
        if (i > 0) {
            host_log((uint32_t)(uintptr_t)" ", 1);
        }
        const char* str = JS_ToCString(ctx, argv[i]);
        if (str) {
            host_log((uint32_t)(uintptr_t)str, strlen(str));
            JS_FreeCString(ctx, str);
        }
    }
    host_log((uint32_t)(uintptr_t)"\n", 1);
    return JS_UNDEFINED;
}

__attribute__((export_name("qjs_std_add_console")))
void qjs_std_add_console(uint32_t ctx_ptr) {
    if (!ctx_ptr) return;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    
    JSValue global = JS_GetGlobalObject(ctx);
    
    // Add print function
    JS_SetPropertyStr(ctx, global, "print", 
        JS_NewCFunction(ctx, js_print, "print", 1));
    
    // Add console object with log method
    JSValue console = JS_NewObject(ctx);
    JS_SetPropertyStr(ctx, console, "log", 
        JS_NewCFunction(ctx, js_print, "log", 1));
    JS_SetPropertyStr(ctx, console, "info", 
        JS_NewCFunction(ctx, js_print, "info", 1));
    JS_SetPropertyStr(ctx, console, "warn", 
        JS_NewCFunction(ctx, js_print, "warn", 1));
    JS_SetPropertyStr(ctx, console, "error", 
        JS_NewCFunction(ctx, js_print, "error", 1));
    JS_SetPropertyStr(ctx, console, "debug", 
        JS_NewCFunction(ctx, js_print, "debug", 1));
    JS_SetPropertyStr(ctx, global, "console", console);
    
    JS_FreeValue(ctx, global);
}

// ============================================================================
// C Function Binding (for Go callbacks)
// ============================================================================

static JSValue go_callback_wrapper(JSContext *ctx, JSValue this_val, 
                                    int argc, JSValue *argv, int magic, 
                                    JSValue *func_data) {
    // func_data[0] contains our callback ID
    int32_t func_id;
    JS_ToInt32(ctx, &func_id, func_data[0]);
    
    // Store arguments as pointers
    uint32_t* arg_ptrs = NULL;
    if (argc > 0) {
        arg_ptrs = (uint32_t*)arena_alloc(sizeof(uint32_t) * argc);
        if (!arg_ptrs) return JS_EXCEPTION;
        for (int i = 0; i < argc; i++) {
            arg_ptrs[i] = store_jsvalue(JS_DupValue(ctx, argv[i]));
        }
    }
    
    // Call the Go callback
    uint32_t result_ptr = host_call_go(
        (uint32_t)(uintptr_t)ctx, 
        func_id, 
        argc, 
        (uint32_t)(uintptr_t)arg_ptrs
    );
    
    return load_jsvalue(result_ptr);
}

__attribute__((export_name("qjs_new_c_function")))
uint32_t qjs_new_c_function(uint32_t ctx_ptr, uint32_t func_id, uint32_t name_ptr, int32_t arg_count) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    const char* name = name_ptr ? (const char*)(uintptr_t)name_ptr : "";
    
    // Store function ID in func_data
    JSValue func_data[1];
    func_data[0] = JS_NewInt32(ctx, func_id);
    
    JSValue func = JS_NewCFunctionData(ctx, go_callback_wrapper, arg_count, 0, 1, func_data);
    
    // Set function name
    if (name[0]) {
        JS_DefinePropertyValueStr(ctx, func, "name", 
                                  JS_NewString(ctx, name), JS_PROP_CONFIGURABLE);
    }
    
    return store_jsvalue(func);
}

// ============================================================================
// Strict Equality
// ============================================================================

__attribute__((export_name("qjs_strict_eq")))
int32_t qjs_strict_eq(uint32_t val1_ptr, uint32_t val2_ptr) {
    JSValue v1 = load_jsvalue(val1_ptr);
    JSValue v2 = load_jsvalue(val2_ptr);
    
    // Use tag and value comparison for strict equality
    return JS_VALUE_GET_TAG(v1) == JS_VALUE_GET_TAG(v2) && 
           JS_VALUE_GET_PTR(v1) == JS_VALUE_GET_PTR(v2);
}

// ============================================================================
// Runtime Configuration
// ============================================================================

__attribute__((export_name("qjs_set_memory_limit")))
void qjs_set_memory_limit(uint32_t rt_ptr, uint32_t limit) {
    if (!rt_ptr) return;
    JSRuntime* rt = (JSRuntime*)(uintptr_t)rt_ptr;
    JS_SetMemoryLimit(rt, limit);
}

__attribute__((export_name("qjs_set_max_stack_size")))
void qjs_set_max_stack_size(uint32_t rt_ptr, uint32_t stack_size) {
    if (!rt_ptr) return;
    JSRuntime* rt = (JSRuntime*)(uintptr_t)rt_ptr;
    JS_SetMaxStackSize(rt, stack_size);
}

// ============================================================================
// Utility: Get Error Message
// ============================================================================

__attribute__((export_name("qjs_get_error_message")))
uint32_t qjs_get_error_message(uint32_t ctx_ptr, uint32_t err_ptr, uint32_t buf_ptr, uint32_t buf_len) {
    if (!ctx_ptr || !buf_ptr || buf_len == 0) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue err = load_jsvalue(err_ptr);
    char* buf = (char*)(uintptr_t)buf_ptr;
    
    // Get error message
    JSValue msg_val = JS_GetPropertyStr(ctx, err, "message");
    const char* msg = JS_ToCString(ctx, msg_val);
    JS_FreeValue(ctx, msg_val);
    
    if (!msg) {
        // Fallback to string representation
        msg = JS_ToCString(ctx, err);
    }
    
    if (!msg) {
        buf[0] = '\0';
        return 0;
    }
    
    size_t msg_len = strlen(msg);
    if (msg_len >= buf_len) msg_len = buf_len - 1;
    memcpy(buf, msg, msg_len);
    buf[msg_len] = '\0';
    
    JS_FreeCString(ctx, msg);
    return msg_len;
}

// Get error with stack trace
__attribute__((export_name("qjs_get_error_stack")))
uint32_t qjs_get_error_stack(uint32_t ctx_ptr, uint32_t err_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    JSValue err = load_jsvalue(err_ptr);
    
    // Try to get stack property
    JSValue stack = JS_GetPropertyStr(ctx, err, "stack");
    if (!JS_IsUndefined(stack) && !JS_IsException(stack)) {
        return store_jsvalue(stack);
    }
    JS_FreeValue(ctx, stack);
    
    // Fallback to toString
    JSValue str = JS_ToString(ctx, err);
    return store_jsvalue(str);
}

// ============================================================================
// Value to String (for debugging/display)
// ============================================================================

__attribute__((export_name("qjs_to_string")))
uint32_t qjs_to_string(uint32_t ctx_ptr, uint32_t val_ptr) {
    if (!ctx_ptr) return 0;
    JSContext* ctx = (JSContext*)(uintptr_t)ctx_ptr;
    return store_jsvalue(JS_ToString(ctx, load_jsvalue(val_ptr)));
}
