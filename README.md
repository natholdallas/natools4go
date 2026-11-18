<div align="center">

# 🧰 natools4go

**A pragmatic, generic-first Go toolkit** — batteries included for building
production applications with **Fiber v3**, **GORM**, **Viper**, and more.

![Go Version](https://img.shields.io/badge/Go-%3E%3D1.25-00ADD8?logo=go&logoColor=white)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Coverage](https://img.shields.io/badge/coverage-core%20packages-green)
[![Go Report Card](https://goreportcard.com/badge/github.com/natholdallas/natools4go)](https://goreportcard.com/report/github.com/natholdallas/natools4go)

Go 开发者的一站式工具箱 · A one-stop toolkit for Go developers

</div>

---

## 📦 Installation

```bash
go get -u github.com/natholdallas/natools4go
```

Requires **Go 1.25+** (uses generics and modern stdlib features).

---

## ✨ Highlights

| 💡 Feature            | 📚 Where                                        |
| :-------------------- | :---------------------------------------------- |
| Generic query builder | `orms.Query[T]` — fluent, type-safe, chainable |
| Generic DB models     | `orms.SoftModel[T]`, `orm.Model[T]`, `IDModel[T]` |
| Fiber request binding | `fext.BodyVarser[T]` + automatic validation     |
| Unified error response| `fext.ErrorHandler` / `fext.Fail`               |
| JWT middleware        | `fext.JWTware`, `fext.GenToken`, `fext.ParseToken` |
| Config management     | `vipers` — typed getters + hot-reload events    |
| Structured validation | `va.Struct` — pretty, readable field errors     |

---

## 🗂 Package Overview

| Package      | Description                                                              | Highlights |
| :----------- | :----------------------------------------------------------------------- | :--------- |
| `slice`      | Generic slice utilities                                                  | `Map`, `Filter`, `ForEach`, `Defu` |
| `maths`      | Integer math helpers                                                     | `DivCeil`, `Digits` |
| `strs`       | String helpers & character constants                                     | `Wrap`, `ToStart/ToEnd`, `Trim*`, `FormatInt/Uint/Float/Bool` |
| `rands`      | Randomization & distribution                                             | `Crypto`/`Char` (CSPRNG), `Hex`, `CryptoDigits`, `Pick`, `Distribute(Strict)`, `Digits` |
| `jsons`      | JSON marshal / nested-map traversal                                      | `Unmarshal[T]`, `IString`, `GetOK`, `Set` |
| `va`         | Validator wrapper (`go-playground/validator/v10`)                        | `Struct`, `Errors` (structured, value-free), `Var` |
| `constraints`| Generic type constraints                                                 | `Signed`, `Unsigned`, `Integer`, `Float` |
| `concur`     | Concurrency primitives                                                   | `Go/Run`, `Pool` (bounded), panic recovery |
| `ask`        | Interactive CLI prompts                                                  | `Read[T]`, `Line`, `ReadPassword` (no echo), `Confirm` |
| `dbg`        | Pretty debugging output                                                  | `JSON`, `Struct`, `Err`, `File` |
| `structs`    | Map / struct conversion (`mapstructure`)                                 | `Map`, `To[T]`, `Vo[T]` (validated) |
| `flags`      | CLI flag helpers                                                         | `Run` |
| `vipers`     | Typed Viper config access with watch events                              | `Get[T]`, `Watch`, `Reload`, `RegisterUpdateHandler` |
| `fext`       | **Fiber v3** extensions: binding, JWT, errors, cache, logging            | `BodyVarser[T]`, `JWTware`, `ErrorHandler`, `Cache` |
| `orms`       | **GORM** wrapper: models, pagination, sorting, typed queries             | `Query[T]`, `Paginate`, `SoftModel[T]`, `SyncTablesSoft/Hard`, `DeleteByID` |
| `narder`     | Tiny leveled logger shared across the toolkits                          | `Info/Infof`, `Warn/Warnf`, `Error/Errorf`, `SetPrinter` |

---

## 🚀 Quick Start

### HTTP service with Fiber + GORM

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/natholdallas/natools4go/fext"
	"github.com/natholdallas/natools4go/orms"
)

type User struct {
	orms.IDModel[int]
	Name string `json:"name" validate:"required"`
}

type CreateUser struct {
	Name string `json:"name" validate:"required"`
}

func main() {
	db := orms.New(postgres.Open("host=localhost user=postgres dbname=app"))
	db.AutoMigrate(&User{})

	app := fiber.New(fiber.Config{ErrorHandler: fext.ErrorHandler})

	app.Post("/users", func(c fiber.Ctx) error {
		body, err := fext.BodyVarser[CreateUser](c) // bind + validate
		if err != nil {
			return &fext.Fail{Status: 400, Message: "bad request", System: err}
		}
		user := User{Name: body.Name}
		if err := orms.Create(db, &user); err != nil {
			return err
		}
		return fext.JSON(c, 201, user)
	})

	fext.Listen(app, ":8080")
}
```

### Typed, fluent queries

```go
type User struct {
	orms.SoftModel[int]
	Email string
	Age   int
}

// Chainable builder, fully type-safe
q := orms.QE[User](db)
user, err := q.Where("age > ?", 18).Order("age desc").First()
// => (User, error)

list, err := orms.QE[User](db).
	Where("email LIKE ?", "%@example.com").
	Preload("Orders").
	Find()
// => ([]User, error)
```

### Pagination — no side effects, no boilerplate

```go
page, tx := orms.Paginate[User](
	orms.QE[User](db).Where("active = ?", true),
	orms.Pagination{Page: 2, Size: 20},
)
// Page{Total: 137, Page: 7, Content: []User{...}}
```

### JWT auth in three lines

```go
app.Use(fext.JWTware("my-secret")) // protects all following routes

token, _ := fext.GenToken("user-42", "my-secret") // 30d default
claims, err := fext.ParseToken(token, "my-secret") // HMAC enforced
```

### Config with hot reload

```go
vipers.Config("config", "./config") // config.toml by default
vipers.Watch(func(e fsnotify.Event) {
	log.Println("config changed:", e.Name)
})

port := vipers.Int("server.port", 8080)  // typed + default
debug := vipers.Bool("server.debug", false)
```

### Validation errors you can actually read

```go
type SignUp struct {
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=18"`
}

err := va.Struct(SignUp{Email: "x"})
// [Email:x]:[required]
// [Age:0]:[gte-18]
```

Values of sensitive fields (names matching `password`, `secret`, `token`,
`apikey`, `hash`, `salt`, …) are redacted as `***`, because these messages are
routinely returned to HTTP clients. For a JSON-friendly, entirely value-free
form use `va.Errors`, which returns `[]FieldError{{Field, Tag, Param}}`.

---

## 🧩 Detailed Guides

### `orms` — GORM wrapper

- **Init**: `orms.New(dialector, opts...)` opens the pool with sensible defaults
  (max idle 10 / max open 100 / lifetime 30m), returning an error on failure.
- **DSN builder**: `orms.DSN(driver, name, user, pass, host, port)` and
  `orms.Dialector(driver, dsn, name, query, prepare...)` cover MySQL / PostgreSQL /
  SQLite / SQL Server / ClickHouse. `prepare=true` auto-creates the database.
- **Models**: `IDModel[T]`, `Model[T]`, `SoftModel[T]` — ready-made generic base models
  with swagger-compatible tags.
- **Query builder**: `orms.Q[T]` (no model), `orms.QE[T]` (model = T),
  `orms.QM[T,M]` (result T, model M), `orms.QT[T](db, table)`.
  Every chainable method on `Query[T]` mirrors GORM and stays **type-safe**.
  Finishers return `(T, error)` or `[]T, error`; `I`-prefixed variants swallow errors
  and return zero values (`IFind`, `IFirst`, `IPaginate` …).
- **Pagination**: `Pagination.Scope` is side-effect free (validates on a copy),
  `Paginate`/`PaginateMapping` give you `Page[T]{Total, Page, Content}`.
- **Sorting**: `Sorter` / `Sorters` implement `Scoper` for safe column ordering.
- **Destructive helpers**: `Delete[T](tx, conds...)` refuses raw string conditions
  (a bare `"1 or 1=1"` becomes a WHERE clause in GORM — the classic route-parameter
  injection). Use `DeleteByID[T](tx, id)` or `DeleteWhere[T](tx, query, args...)`.
- **Schema reconciliation**: `SyncTablesSoft` adds missing columns and keeps
  database-only ones (moved last, sorted); `SyncTablesHard` also drops columns that
  no longer exist in the struct. Both process **every** model passed in. Prefer
  `AutoMigrate` for routine add-only migrations.
- **Reset**: `Reset(db, driver, dsn)` drops and recreates the whole database;
  `AutoCreate`/`AutoCreateE` create it when missing. Both accept GORM driver names
  and translate them to the matching `database/sql` driver. SQLite needs neither —
  file databases create themselves, so those calls are no-ops.
- **JSON columns**: `List[T]` / `Dict[T]` implement `sql.Scanner` + `driver.Valuer`
  for JSON persistence across databases.
- **Logging**: `orms.LogPreset(out, level)` builds a stdlib-backed GORM logger.

### `fext` — Fiber v3 extensions

- **Binding**: `BodyParser`, `BodyVarser`, `QueryParser/Varser`, `ParamsParser/Varser`,
  `RestParser/Varser`, `CookieParser/Varser`, `ReqHeaderParser/Varser`, `FormData`.
  The `Varser` suffix = **bind + validate** (`va.Struct`). Structs may implement
  `Initializer` to run after binding.
- **Errors**: `fext.Fail{Status, Code, Message, System}` — returns a clean JSON body.
  `ErrorHandler` maps `*Fail`, `*fiber.Error`, and generic errors to proper status
  codes; `System` is surfaced only in debug mode (`narder.SetDebugMode(true)`) and forwarded
  to `SetErrorFunc`.
- **JWT**: `JWTware` (HS256 middleware), `GenToken` / `ParseToken` (algorithm-confusion
  safe — only HMAC accepted), `JWT` struct for grouped secret management,
  `JWT.Claims(c)` to read the current user (nil when unauthenticated).
- **Utils**: `Cache(seconds)`, `Status`/`JSON`/`SendString`, `GetAuthorization`,
  and `Listen` (logs startup failures via `narder`, forwards errors to
  `SetErrorFunc`, or panics on startup failure).

### `vipers` — typed config

Every getter accepts an optional default: `Get[T]`, `String`, `Bool`, `Int*`, `Uint*`,
`Float64`, `Time`, `Duration`, `IntSlice`, `StringSlice`, `StringMap*`, `SizeInBytes`.
`Watch` subscribes to file changes; handlers registered via `RegisterUpdateHandler` are
dispatched under a lock (`Reload`), each invoked in registration order on every
config change.

### `concur` — concurrency done right

- `Go(tasks...)` / `Run(tasks...)` — fan out and join.
- `Pool{workers}` — bounded concurrency via `NewPool(n)`, `Submit`, `Wait`, `Close`.
  Submitting after `Close` panics to catch misuse.
- Panics in tasks are recovered, logged through `narder` at ERROR level and
  **swallowed** — a panicking task cannot take the process down (a panic inside a
  goroutine cannot be recovered by the caller anyway). Use `SetPanicHandler` to
  escalate, e.g. re-panic when a fail-fast startup is preferred.

### `jsons` — JSON helpers

`Unmarshal[T]`, `Marshal(v, pretty?)`, `String(v, pretty?)`, plus `I`-variants that
ignore errors. `Map(v)` converts any value to `map[string]any`.
`Get`/`Set`/`GetOK` traverse nested maps **without panicking** on missing keys.

### `dbg` — debug printing

`JSON`, `Struct`, `Err`, `Dump`, and `File` (cat-like). Output goes through the
shared `narder` logger; call sites are resolved to the caller of these
functions, not to internal library frames.

### `narder` — shared leveled logger

A minimal `Info`/`Warn`/`Error` logger used by `dbg`, `fext`, and `vipers`.

```go
narder.Infof("user %d signed in", id)
narder.Error("db unavailable")

// Route everything through your own backend
narder.SetPrinter(func(level, file string, line int, msg string) {
    slog.Info(msg, "level", level, "file", file, "line", line)
})

// Or tune the stdlib backend (default: stderr, Ldate|Ltime)
narder.SetOutput(file)
narder.SetFlags(log.Lshortfile) // file:line of YOUR call site
```

Call-site resolution skips every frame inside the natools4go module, so even
wrapped calls (e.g. `dbg.Dump`) report the caller's position — not the library
internals.

---

## 🧪 Testing

The core pure-function packages (`slice`, `maths`, `strs`, `rands`, `jsons`, `concur`)
ship with table-driven tests:

```bash
go test ./...      # run all tests
go test -race ./... # race detector for concurrent packages
go vet ./...
```

---

## 🧭 Project Layout

```
natools4go/
├── ask/          # interactive CLI prompts
├── concur/       # goroutines, bounded pool, panic recovery
├── constraints/  # generic type constraints
├── fext/         # Fiber v3: binders, JWT, error handling, cache
├── flags/        # CLI flag helpers
├── jsons/        # JSON marshal + nested map helpers
├── maths/        # integer math
├── narder/       # tiny leveled logger (Info/Warn/Error)
├── orms/         # GORM: models, query builder, pagination
├── rands/        # random strings, distribution, shuffle
├── slice/        # slice utilities
├── dbg/          # pretty debugging
├── strs/         # string helpers & constants
├── structs/      # mapstructure conversion
├── va/           # validator wrapper
└── vipers/       # typed Viper config + hot reload
```

---

## 📄 License

[MIT](LICENSE) © natools4go contributors.