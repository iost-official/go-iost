# Agent Guide for go-iost

This file contains practical guidance for AI coding agents working on the `go-iost` repository.

## Project Overview

`go-iost` is the Go implementation of the IOST blockchain node. It provides:

- A full PoB (Proof-of-Believability) consensus node (`iserver`)
- A CLI wallet/client (`iwallet`)
- An integration test runner (`itest`)
- A JavaScript smart-contract execution environment

Module path: `github.com/iost-official/go-iost/v3`

> **Note:** The top-level `README.md` is currently outdated. It still references the old C++ V8 JavaScript engine and Go 1.20. The current codebase uses `dop251/goja` (pure Go) and requires Go 1.25+. Treat this file as the authoritative source for development setup.

## Environment Requirements

- **Go:** 1.25 or later (required by `dop251/goja`)
- **CGO:** Disabled (`CGO_ENABLED=0` is set in the `Makefile`)
- **Architecture:** Historically amd64-only because of the old V8 native dependency. With the migration to goja, arm64/darwin builds work as well, but the `Makefile` still forces `GOARCH=amd64` for release builds.
- **OS:** Linux and macOS are both used in CI (`ubuntu-22.04` and `macos-latest`).

## Quick Build

```bash
# Build all three binaries into target/
make build

# Individual binaries
make iserver
make iwallet
make itest
```

Built artifacts land in `target/`.

## Running Tests

### Unit / core tests

```bash
# Fast, broad test run (matches Linux_Test core packages)
go test -timeout 600s -p 1 -count=1 ./account/... ./chainbase/... ./common/... \
  ./consensus/... ./core/... ./crypto/... ./db/... ./ilog/... ./metrics/... \
  ./p2p/... ./rpc/... ./sdk/... ./vm/v8vm/...

# VM-specific tests live under test/v8vm
go test -timeout 600s -p 1 -count=1 ./test/v8vm/...

# Integration tests
go test -timeout 600s -p 1 -count=1 ./test/integration/...
```

Use `-p 1` because several tests bind fixed ports or share on-disk state.

### Full Makefile test

```bash
make test
```

### Lint

```bash
make lint-tool   # installs latest golangci-lint
make lint        # runs .golangci.yml
```

## Local Chain Smoke Test

A quick end-to-end sanity check:

```bash
# 1. Build
make build

# 2. Start a local node (use a temp directory to avoid dirtying the repo)
mkdir -p /tmp/iost-local/config
cp -r config/genesis /tmp/iost-local/config/
# copy config/iserver.yml and adjust paths to /tmp/iost-local
./target/iserver -f /tmp/iost-local/config/iserver-local.yml

# 3. In another terminal, import the genesis admin key and create an account
./target/iwallet --server localhost:30002 --chain_id 1020 \
  account import admin 2yquS3ySrGWPEKywCPzX4RTJugqRh7kJSo5aehsLYPEWkUxBWA39oMrZ7ZxuM4fgyXYs2cPwh5n8aNNpH5x2VyK1

./target/iwallet --server localhost:30002 --chain_id 1020 \
  account create testuser --account admin --initial_balance 100

# 4. Send a transfer
./target/iwallet --server localhost:30002 --chain_id 1020 \
  transfer admin 1 --account testuser --memo "smoke test"
```

Default RPC ports from `config/iserver.yml`:

- REST gateway: `30001`
- gRPC: `30002`
- Debug/pprof: `30003`
- P2P: `30000`

Default `chain_id` for the local dev config is `1020` (iwallet defaults to `1024`, so pass `--chain_id 1020`).

## Architecture at a Glance

| Directory | Purpose |
|---|---|
| `cmd/iserver`, `cmd/iwallet`, `cmd/itest` | CLI entry points |
| `vm/v8vm` | JavaScript smart-contract runtime (now goja-based) |
| `vm/host` | Host environment exposed to contracts (storage, blockchain context, gas) |
| `core/contract`, `core/tx`, `core/block` | Core blockchain data structures |
| `consensus/pob` | Proof-of-Believability consensus |
| `chainbase` | Chain state, block processing, recovery |
| `p2p` | libp2p networking |
| `rpc` | gRPC / REST gateway |
| `config/genesis` | Genesis contracts and configuration |
| `test/v8vm`, `test/integration`, `test/native`, `test/gas` | Test suites |

## JavaScript VM (`vm/v8vm`)

The old C++ V8 engine and the intermediate QuickJS/wazero implementation have been removed. The current implementation uses `github.com/dop251/goja`.

Key files:

- `vm/v8vm/sandbox.go` — creates a goja runtime, loads runtime libs, executes contract code with gas/deadline enforcement.
- `vm/v8vm/host_bindings.go` — Go callbacks exposed to JS (`IOSTBlockchain`, `IOSTStorage`, `IOSTInstruction`, `_IOSTCrypto`).
- `vm/v8vm/libjs/*.js` — embedded JS runtime libraries (`//go:embed`).
- `vm/v8vm/pool.go` — VM pool for compile and run sandboxes.
- `vm/monitor.go` — factory that creates the VM pool (`v8.NewVMPool(10, 400)`).

Important implementation details:

- Runtime libraries are embedded at build time; `vm.jspath` in `config/iserver.yml` is obsolete and ignored.
- Gas is enforced in `_IOSTInstruction_counter.incr()` inside `host_bindings.go`.
- Execution deadline is enforced via `goja.Runtime.Interrupt()` triggered by `time.AfterFunc`.
- The result length limit is measured in UTF-16 code units to match JS `String.prototype.length`.
- There is no native per-runtime memory limit; pathological allocations are mitigated by wrapping the `Array` constructor.

When changing JS runtime behavior, test with:

```bash
go test ./test/v8vm/... ./test/integration/...
```

## CI / CD

Workflow file: `.github/workflows/master-actions.yml`

Jobs:

- `Mac_Test` — build, lint, full `make test` on macOS
- `Linux_Test` — build, lint, then package-level test matrix on Ubuntu
- `E2E_Test_Local` — builds Docker image and runs `itest` cases

All jobs currently use Go 1.25.

## Coding Conventions

- Run `gofmt -s -w` on changed Go files; `make format` formats the whole tree.
- Follow existing package naming; the JS VM package is still called `v8vm` for historical reasons.
- Avoid introducing new CGO dependencies. The project is intentionally CGO-free now.
- Keep changes minimal. Prefer targeted fixes over large refactors.
- If a change touches JS-host bindings or gas accounting, add/update tests in `test/v8vm` or `test/integration`.

## Common Tasks

### Regenerate protobuf

```bash
make protobuf
```

This runs `script/gen_protobuf.sh`. The old V8 library exports were removed; it is now a standard protobuf generation script.

### Build Docker image

```bash
make image
```

Uses `Dockerfile.run` (production-style image). `Dockerfile` is for release builds.

### Clean local dev artifacts

```bash
make clear_debug_file
```

Removes `storage/`, `logs/`, `ilog/logs*`, and p2p peer files created by `make debug`.

## Things to Watch Out For

- **Fixed ports:** Many tests and `make debug` use hard-coded ports `30000–30003`. Running multiple instances on the same machine without port changes will fail.
- **Test ordering / shared state:** Some integration tests rely on the same `storage/` directory; use `-p 1` and clean between runs when flaky behavior appears.
- **goja version:** Pin to the commit currently in `go.mod`. goja updates can change JavaScript semantics and gas costs.
- **Workflow file changes:** Pushing `.github/workflows/*.yml` via the HTTPS remote may be rejected by GitHub for OAuth apps lacking the `workflow` scope. If that happens, use the SSH remote (`git@github.com:iost-official/go-iost.git`).

## Useful Commands Reference

```bash
make build              # build iserver/iwallet/itest
make test               # run all Go tests
make lint               # run golangci-lint
make debug              # build + run local iserver with config/iserver.yml
make e2e_test_local     # build + run a_case/t_case/c_case locally
make image              # build Docker image via Dockerfile.run
make protobuf           # regenerate protobuf/GRPC bindings
make clean              # remove target/
make clear_debug_file   # remove local dev storage/logs
```
