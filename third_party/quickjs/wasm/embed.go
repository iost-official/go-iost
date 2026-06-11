// Package wasm embeds the QuickJS-ng WebAssembly binary.
package wasm

import _ "embed"

// QuickJS is the embedded QuickJS-ng WebAssembly binary.
//
//go:embed quickjs.wasm
var QuickJS []byte
