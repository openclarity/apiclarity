package rest

import (
	 _ "embed"
)

//go:embed http-trace-filter.wasm
var wasmPluginBinary []byte
