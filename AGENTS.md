# Repository Guidelines

`github.com/natholdallas/natools4go` is a **pure Go library** (no `main`, no binaries, no mocks/generated code): a generic-first toolkit for Fiber v3 HTTP services and GORM-backed applications. Packages are independent; consumers import only what they need.

## Project Overview

| Package | Purpose |
| --- | --- |
| `slice` | `Map`, `Filter`, `ForEach`; `Defu` (first arg or default, the module-wide optional-arg helper) |
| `constraints` | `Signed`, `Unsigned`, `Integer`, `Float` type sets |
| `maths` | `DivCeil`, `Digits` |
| `strs` | Char constants, `Wrap`/`ToStart`/`Trim*`, `Format*`/`Parse*` conversions |
| `rands` | CSPRNG strings (`Crypto`, `Hex`, `CryptoDigits`), `CharFast` (insecure), `Pick`, `Distribute`, shuffle |
| `jsons` | Generic marshal/unmarshal, nested `map[string]any` traversal (`Get/GetOK/Set`) |
| `structs` | `Map`, `To` (silent), `ToE`, `Vo` (decode + validate) |
| `va` | validator/v10 wrapper: `Struct`, `Errors`, `Var`, value redaction |
| `concur` | `Go`/`Run`, bounded `Pool`, panic recovery + `SetPanicHandler` |
| `ask` | Stdin prompts (`Read[T]`, `Line`, `ReadPassword`, `Confirm`) |
| `dbg` | Debug printing through `narder` (`JSON`, `Struct`, `Err`, `File`) |
| `flags` | `Run(cond, fn)` conditional launcher; exits the process |
| `narder` | Shared leveled logger (`Info/Warn/Error`), call-site resolution |
| `vipers` | Typed Viper getters + default values, hot-reload handlers |
| `fext` | Fiber v3 binding/validation, `Fail` error envelope, JWT, cache, listen |
| `orms` | GORM wrapper: dialects, DSNs, migrations, generic models, `Query[T]`, pagination |
| `pwd` | Argon2id `Hash`/`Verify`/`IsHashed` (not listed in README) |

## Architecture & Data Flow

- **Layering.** Leaf utilities (`slice`, `maths`, `strs`, `constraints`, `jsons`) have no intra-module deps. `narder` is the logging spine used by `concur`, `dbg`, `fext`, `orms`. `va` is the shared validator used by `fext`, `vipers`, `structs`. `fext` and `orms` are the two application-facing packages.
- **No DI container.** Dependencies are explicit parameters; mutable process-wide state lives in `narder` (backend/hooks), `va.validate`, and Viper globals. There is no config struct for `fext`.
- **orms connection flow:** `DSN` builds a database-less DSN → `Dialector` injects the DB name/query params (and creates the database when `prepare=true`) → `New` opens GORM and sets pool limits (SQLite is pinned to 1 connection) → `AutoMigrate` / `SyncTablesSoft|Hard` / `ResetTables`.
- **orms query flow:** `Q` (no model) / `QE` (model = `T`) / `QM` / `QT` → chainable methods mutate `q.db` and return the same `*Query[T]` → finishers (`First`, `Find`, `Paginate`, `Exec`, …) run GORM and convert results. `I*` finishers swallow errors. `Pagination`, `Sorter`, `Sorters` implement `Scope(*gorm.DB) *gorm.DB`.
- **fext request flow:** `c.Bind().URI/Query/Body/Cookie/Header` chain via an internal `binder` → `TryInitialize` (`Initializer` hook) → optional `va.Struct` for `*Varser` helpers → handlers return `*fext.Fail` or any error → `fext.ErrorHandler` renders the JSON envelope. Success responses are raw data; there is no success wrapper.
- **DDL reconciliation:** `SyncTablesSoft` adds columns and keeps/moves extras last; `SyncTablesHard` also drops extras; both reorder only when needed and emit zero DDL when already ordered. Dialect-specific rebuilds live in `orms/rebuild.go`.

## Key Directories

| Path | Contents |
| --- | --- |
| `ask/`, `concur/`, `dbg/`, `flags/`, `jsons/`, `maths/`, `pwd/`, `rands/`, `slice/`, `strs/`, `structs/` | Small, single-purpose utility packages |
| `constraints/` | Type-set definitions only |
| `narder/` | Logger plus caller-frame skipping; changing the module path breaks it |
| `vipers/` | Global Viper config access + fsnotify reload |
| `fext/` | Fiber v3 helpers: `fext.go` binders, `fail.go` errors, `jwt.go` auth, `cache.go`, `ctx.go` |
| `orms/` | `dialect.go` DSNs/lifecycle, `queries*.go` builder, `extensions.go` models/pagination, `rebuild.go` DDL, `datatypes.go` JSON columns |
| `orms/testdata/` | Docker Compose services + MySQL grant script for integration tests |

## Development Commands

```bash
go build ./...                       # compile all packages
go vet ./...
go test ./...                        # unit tests (in-memory SQLite)
go test -race -count=1 ./...         # CI parity
gofmt -l .                           # CI gate: must print nothing
gofumpt -l .                         # repo convention (not in CI)
gci diff -s standard -s default -s 'Prefix(github.com/natholdallas/natools4go)' .
golangci-lint run                    # same config as CI
```

Run a single package/test: `go test ./orms/ -run TestSyncTablesSoft -v`.

Integration tests (never run in CI):

```bash
docker compose -f orms/testdata/docker-compose.yml up -d
export ORMS_MYSQL_DSN=... ORMS_POSTGRES_DSN=... ORMS_SQLSERVER_DSN=... ORMS_CLICKHOUSE_DSN=...
go test -tags integration ./orms/ -run 'TestSyncTablesEveryDriver|TestResetEveryDriver' -v
```

Unset DSNs are skipped individually; with none set the test calls `t.Skip`. Compose ports/credentials are in `orms/testdata/docker-compose.yml` (MySQL is published on `13306`).

## Code Conventions & Common Patterns

- **Formatting:** `gofmt` + `gofumpt`; tabs (width 4). Imports are grouped stdlib / third-party / `github.com/natholdallas/natools4go` (blank line before the local group). The `.golangci.yml` comment's `Prefix(webtplmst)` is stale — use the module prefix above.
- **Naming:** `MustX` panics (`MustNew`, `MustAutoMigrate`, `MustStruct`, `MustConfig`); `IX`/`TryX` swallow errors and return zero values (`IFind`, `IString`, `TryHash`); `XE` returns an error (`ToE`, `ParseIntE`, `AutoCreateE`). In `fext`, `*Parser` binds only and `*Varser` binds + validates.
- **Errors:** return errors, wrap with `fmt.Errorf("context: %w", err)`; libraries do not call `log.Fatal`/`os.Exit` (the exception is `flags.Run` by design). Log-and-return is avoided; `dbg` helpers log because they cannot return errors.
- **Generics over reflection:** typed wrappers (`Query[T]`, `Read[T]`, `List[T]`, `Page[T]`); reflection is only indirect via `encoding/json`, mapstructure, validator, Viper. Use `any`, never `interface{}` (revive `use-any`).
- **Optional arguments:** trailing variadics (`pretty ...bool`, `endtime ...time.Duration`, `prepare ...bool`), resolved with `slice.Defu`.
- **Concurrency:** use `concur.Go`/`Run` and `Pool`; call `Wait` before `Close`. Panics in tasks are recovered, logged, and swallowed unless `SetPanicHandler` escalates. SQLite's single-connection pool means do not query the same `*gorm.DB` concurrently with an open transaction.
- **Security invariants (do not weaken):** `orms.Delete` rejects non-numeric string conditions (`DeleteByID`/`DeleteWhere` are the safe paths); `Sorter.Valid` allows bare identifiers only; DB names go through `validateDBName` + `quoteIdent`; `va` redacts credentials as `***`; JWT parsing accepts HMAC only and requires expiry; `fext.Fail.System` is serialized only when `narder.IsDebug()`, while unknown errors and `*fiber.Error` messages at status ≥500 are masked.
- **Global hooks:** configure `narder.SetPrinter`/`SetDebugMode`, `fext.SetErrorHandler`, and `vipers.Watch` (first call only) at startup; `vipers.RegisterUpdateHandler` adds later handlers.

## Important Files

| File | Why it matters |
| --- | --- |
| `go.mod` | Go 1.25.0 baseline; Fiber v3, GORM + 5 drivers, Viper, validator |
| `.github/workflows/ci.yml` | Format, build, vet, race test, lint gates |
| `.golangci.yml` | Analysis linters only: govet (all minus `fieldalignment`), staticcheck all, revive `use-any`/`var-naming`, unused, misspell |
| `orms/orms.go` | `New`, `AutoMigrate`, sync entry points, `ResetTables` |
| `orms/dialect.go` | DSN/dialector construction, DB create/reset per driver |
| `orms/rebuild.go` | Per-dialect column reordering/table rebuild |
| `orms/func.go` | Generic CRUD; delete-condition guard |
| `orms/extensions.go` | Models, `Pagination`, `Sorter(s)`, `Paginate`, `LogPreset` |
| `fext/fext.go` | Binder chain; all `Parser`/`Varser` helpers |
| `fext/fail.go` | `Fail` envelope + `ErrorHandler` |
| `fext/jwt.go` | HMAC-only token issuance/parsing, middleware |
| `narder/narder.go`, `va/va.go`, `vipers/vipers.go` | Hold process-wide state: logger backend/debug flag, validator singleton, global Viper instance |

## Runtime/Tooling Preferences

- **Go only.** Required version is `go 1.25.0` from `go.mod`; CI installs exactly that via `go-version-file`. No Node/Bun/package manager, no code generation.
- **cgo is expected:** `gorm.io/driver/sqlite` uses `mattn/go-sqlite3`, and `go test -race` requires cgo. CI runs on Ubuntu only, so platform-specific behavior is not validated.
- **Lint is strict:** `default: none` plus explicit linters, `max-issues-per-linter: 0`, so every finding fails CI. Fix code rather than adding suppressions; `golangci-lint` runs `latest`.
- **SQLite vs real servers:** in-process SQLite is the unit-test default; MySQL/Postgres/SQL Server/ClickHouse behavior is only covered by `//go:build integration` tests.

## Testing & QA

- **Framework:** stdlib `testing` only — no testify or assertion libraries. Assertions are manual `if got != want { t.Errorf(...) }`; helpers such as `newTestDB` (`orms/orms_test.go`), `seedUsers`, `tableColumns`, and `sqlRecorder` live beside the tests.
- **Style:** individual tests per behavior; table-driven only for input matrices (`maths`, `constraints`, `orms/orms_test.go`, `sync_integration_test.go`). No `t.Parallel`, benchmarks, fuzz targets, or `Example*` functions.
- **Isolation:** `t.TempDir()` for files, `t.Cleanup` to restore global state (`narder` backends, `concur.SetPanicHandler`, stdin), and per-test named in-memory SQLite (`file:<t.Name()>?mode=memory&cache=shared`).
- **Integration gating:** compile-time `//go:build integration` in `orms/sync_integration_test.go` plus runtime `os.Getenv` per driver (`ORMS_MYSQL_DSN`, `ORMS_POSTGRES_DSN`, `ORMS_SQLSERVER_DSN`, `ORMS_CLICKHOUSE_DSN`). Never let a DB-dependent test enter the default `go test ./...` path.
- **Coverage:** no coverage tooling or threshold is enforced; the README claims coverage only for core pure-function packages. Add tests for new behavior in the package you touch, using existing patterns; dialect-specific DDL needs the integration tag.
- **CI gates:** `gofmt -l .`, `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...`, then `golangci-lint`. Integration tests, coverage, and `gofumpt`/`gci` checks are not run by CI.
