# Makefile for compiling QuickJS-ng to WebAssembly
#
# Prerequisites:
#   Run: ./setup-wasi-sdk.sh
#   Or:  make WASI_SDK_PATH=/path/to/wasi-sdk
#

# wasi-sdk path
WASI_SDK_PATH ?= $(HOME)/wasi-sdk

# Verify wasi-sdk exists
ifeq ($(wildcard $(WASI_SDK_PATH)/bin/clang),)
$(error wasi-sdk not found at $(WASI_SDK_PATH). Run ./setup-wasi-sdk.sh first)
endif

CC = $(WASI_SDK_PATH)/bin/clang

# Output directories
BUILD_DIR = build
WASM_DIR = ../wasm

# Source directory
QJS_SRC = quickjs-ng

# WASM compilation flags - optimize for size (better for wazero interpretation)
CFLAGS = \
	-Os \
	-flto \
	-D_GNU_SOURCE \
	-DCONFIG_VERSION=\"0.8.0\" \
	-DNDEBUG \
	-I$(QJS_SRC) \
	-I. \
	-fvisibility=hidden

# Linker flags for WASM - reactor mode (no main, exports only)
# Initial memory of 16MB is sufficient for most use cases and reduces startup time
LDFLAGS = \
	-Wl,--no-entry \
	-Wl,--export-dynamic \
	-Wl,--allow-undefined \
	-Wl,--initial-memory=16777216 \
	-Wl,--max-memory=134217728 \
	-Wl,--lto-O3 \
	-Wl,--gc-sections \
	-nostartfiles

# Source files
SRCS = \
	$(QJS_SRC)/quickjs.c \
	$(QJS_SRC)/cutils.c \
	$(QJS_SRC)/dtoa.c \
	$(QJS_SRC)/libunicode.c \
	$(QJS_SRC)/libregexp.c \
	bridge.c

# Main target
all: $(WASM_DIR)/quickjs.wasm

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(WASM_DIR):
	mkdir -p $(WASM_DIR)

# Compile directly to WASM (LTO requires single-step compilation)
$(WASM_DIR)/quickjs.wasm: $(SRCS) | $(WASM_DIR)
	$(CC) $(CFLAGS) $(LDFLAGS) $(SRCS) -o $@
	@echo ""
	@echo "=== Build successful ==="
	@ls -lh $@
	@echo ""
	@echo "Exported functions:"
	@$(WASI_SDK_PATH)/bin/llvm-nm $@ 2>/dev/null | grep " T " | grep "qjs_" | head -20 || true

# Optimize with wasm-opt if available
optimize: $(WASM_DIR)/quickjs.wasm
	@which wasm-opt > /dev/null && \
		wasm-opt -Os --enable-bulk-memory $(WASM_DIR)/quickjs.wasm -o $(WASM_DIR)/quickjs.wasm && \
		echo "Optimized with wasm-opt" && ls -lh $(WASM_DIR)/quickjs.wasm || \
		echo "wasm-opt not found, skipping optimization"

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(WASM_DIR)/quickjs.wasm

# Show info about built WASM
info: $(WASM_DIR)/quickjs.wasm
	@echo "=== WASM Info ==="
	@ls -lh $(WASM_DIR)/quickjs.wasm
	@echo ""
	@echo "Exports:"
	@$(WASI_SDK_PATH)/bin/llvm-nm $(WASM_DIR)/quickjs.wasm 2>/dev/null | grep " T " | grep "qjs_" || true

.PHONY: all clean info optimize
