package streammux

import core "github.com/libp2p/go-libp2p-core/mux"

// Deprecated: use github.com/libp2p/go-libp2p-core/mux.ErrReset instead.
var ErrReset = core.ErrReset

// Deprecated: use github.com/libp2p/go-libp2p-core/mux.MuxedStream instead.
type Stream = core.MuxedStream

// Deprecated: use github.com/libp2p/go-libp2p-core/mux.NoopHandler instead.
var NoOpHandler = core.NoopHandler

// Deprecated: use github.com/libp2p/go-libp2p-core/mux.MuxedConn instead.
type Conn = core.MuxedConn

// Deprecated: use github.com/libp2p/go-libp2p-core/mux.Multiplexer instead.
type Transport = core.Multiplexer
