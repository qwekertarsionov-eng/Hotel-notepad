# 03 — Architecture (Architecture)

**System:** Guest Accounting
**Language:** English
**Pattern:** Railway-Oriented Programming (ROP)

---

## 1. Layer Overview

Architecture is unidirectional: a request passes through layers top to bottom, the result
(success or error) is returned back as `Either`/`IOEither`.

```
┌─────────────────────────────────────────────────────────────┐
│                     HTTP client (Frontend)                  │
└──────────────────────────────┬──────────────────────────────┘
                               │ HTTP / JSON
┌──────────────────────────────▼──────────────────────────────┐
│  Handler (transport)                                        │
│  - routing, JSON parsing, JWT middleware                    │
│  - converts domain (T, error) → HTTP response               │
└──────────────────────────────┬──────────────────────────────┘
                               │ IOEither[DomainError, T]
┌──────────────────────────────▼──────────────────────────────┐
│  Service (application / use-case)                           │
│  - business logic, Either/IOEither composition              │
│  - validation via domain factories                          │
└──────────────────────────────┬──────────────────────────────┘
                               │ IOEither[DomainError, T]
┌──────────────────────────────▼──────────────────────────────┐
│  Repository (port → adapter)                                │
│  - interfaces (port) in service layer                       │
│  - PostgreSQL adapters in infra layer; return (T, error)    │
└──────────────────────────────┬──────────────────────────────┘
                               │ (T, error)  ← standard Go
┌──────────────────────────────▼──────────────────────────────┐
│  DB (PostgreSQL)                                            │
│  - tables users, guests; migrations; indexes; transactions  │
└─────────────────────────────────────────────────────────────┘
```

ASCII scheme of the ROP flow for the "create guest" use-case:

```
  request
    │
    ▼
[decode JSON] ──error──▶ Left(BAD_REQUEST) ─────────────┐
    │ success                                            │
    ▼                                                    │
[auth middleware] ──no token──▶ Left(AUTH_*) ────────────┤
    │ userId                                             │
    ▼                                                    │
[ValidateGuest] ──violation──▶ Left(VALIDATION_FAILED) ──┤
    │ Right(Guest)                                       │
    ▼                                                    │
[repo.Insert] ──DB failure──▶ Left(INTERNAL) ────────────┤
    │ Right(saved)                                       │
    ▼                                                    ▼
Right(saved) ──▶ HTTP 201         Fold → HTTP response (2xx/4xx/5xx)
```

---

## 2. Components and Their Responsibilities

| Layer | Package (suggested) | Responsibility | Result type |
|---|---|---|---|
| Handler | `internal/transport/http` | HTTP parsing, form validation, status mapping | `Either` → HTTP |
| Middleware | `internal/transport/middleware` | JWT check, `user_id` extraction | `Either`/`(Claims, error)` |
| Service | `internal/service` | business logic composition, orchestration | `IOEither` |
| Repository (port) | `internal/repository` | data access interfaces | `(T, error)` |
| Repository (adapter) | `internal/repository/postgres` | SQL query implementation | `(T, error)` |
| Domain | `internal/domain` | entities, factories, validation | `Either` |
| ROP types | `internal/rop` | `Either`, `IOEither` | values |
| Config | `internal/config` | environment reading | `(Config, error)` |

Port/adapter separation: the service layer depends only on interfaces (repository ports); the implementation in the postgres adapter returns standard `(T, error)`, which at the service boundary is wrapped into `IOEither`.

---

## 3. Either[E,T] Contract

`Either[E,T]` — a value of exactly one of two states: `Left[E]` (error) or `Right[T]` (success). Composition methods:

| Method | Signature (descriptive) | Behavior |
|---|---|---|
| `Map` | `Either[E,T].Map(T→U) → Either[E,U]` | applies function to `Right`, passes `Left` as is |
| `FlatMap` | `Either[E,T].FlatMap(T→Either[E,U]) → Either[E,U]` | passes `Right` to the function; `Left` does not call the function |
| `MapLeft` | `Either[E,T].MapLeft(E→F) → Either[F,T]` | applies function to `Left`, passes `Right` |
| `Fold` | `Either[E,T].Fold(E→X, T→X) → X` | folds both states into one value |

Note: in Go, due to the absence of generic methods (methods cannot introduce new type parameters), operations such as `Map(T→U)` are implemented as functions:

```go
func Map[E, T, U any](e Either[E, T], f func(T) U) Either[E, U]
func FlatMap[E, T, U any](e Either[E, T], f func(T) Either[E, U]) Either[E, U]
func MapLeft[E, F, T any](e Either[E, T], f func(E) F) Either[F, T]
func Fold[E, T, X any](e Either[E, T], onErr func(E) X, onOk func(T) X) X
```

This is a Go generic-method limitation: free type functions, not methods.

---

## 4. IOEither[E,T] Contract

`IOEither[E,T]` — a lazy description of an effect that upon execution returns
`Either[E,T]`. The effect is not executed until `Run`.

| Method/function | Signature (descriptive) | Behavior |
|---|---|---|
| `FromIO` | `FromIO(func() (T, error)) → IOEither[E,T]` | wraps effect `(T, error)`; on `error != nil` → `Left` with cause |
| `MapIO` | `IOEither.MapIO(T→U) → IOEither[E,U]` | deferred success transformation |
| `FlatMapIO` | `IOEither.FlatMapIO(T→IOEither[E,U]) → IOEither[E,U]` | deferred composition; does not execute effects until `Run` |
| `Run` | `IOEither.Run(ctx) → Either[E,T]` | executes effect (once), respects `ctx` cancellation |

```go
type IOEither[E, T any] interface {
    MapIO[U any](f func(T) U) IOEither[E, U]
    FlatMapIO[U any](f func(T) IOEither[E, U]) IOEither[E, U]
    Run(ctx context.Context) Either[E, T]
}
```

Applicability boundary: `(T, error)` remains the standard for infrastructure interfaces (repository, HTTP clients). `IOEither` is a composition method, not a reliability guarantee: it does not cancel the need for timeouts, retries, and transactions.

---

## 5. Seven ROP Invariants

| # | Invariant | How it is verified (Test ID) |
|---|---|---|
| 1 | `Either` contains exactly one valid state (not `Left`+`Right` simultaneously, not "empty") | `TST-ROP-001` |
| 2 | `FlatMap(Left)` does not call the next function (call counter = 0) | `TST-ROP-002` |
| 3 | `Map(Left)` preserves the original error | `TST-ROP-003` |
| 4 | `IOEither` does not execute the effect before `Run` (effect counter = 0 before start) | `TST-ROP-004` |
| 5 | `FromIO` does not lose the error cause (`error` → `Left` with same cause) | `TST-ROP-005` |
| 6 | Context cancellation is passed to nested effects (`ctx.Err() != nil` → `Left`) | `TST-ROP-006` |
| 7 | Repeated `Run` may repeat the effect (effect is not memoized) | `TST-ROP-007` |

Note on invariant 7: `IOEither` does not cache the result. Two consecutive `Run`s execute the effect twice. Guarantee of exactly one execution within a single `Run` is a separate property, not memoization between runs.

---

## 6. Typical Use-Case in ROP Style (pseudocode)

```go
// Lower layers return (T, error); service wraps into IOEither.
func (s *GuestService) Create(ctx context.Context, in CreateGuestInput, userId int64) IOEither[DomainError, Guest] {
    return FlatMapIO(
        FromIO(func() (Guest, error) { return decodeInput(in), nil }),   // 1. parsing
        func(g Guest) IOEither[DomainError, Guest] {
            return MapIO(                                                // 2. validation (pure)
                FromIO(func() (Guest, error) { return g, nil }),
                func(g Guest) Guest { ... },
            )
        },
    )
    // idea: validate → insert → return Right(saved)
}

// Handler folds Either into HTTP response:
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    result := h.svc.Create(r.Context(), in, userId).Run(r.Context())
    Fold(result,
        func(e DomainError) { writeError(w, e) },   // Left → code from 06
        func(g Guest)       { writeJSON(w, 201, g) }, // Right → 201
    )
}
```

Actual composition "validation → insert → response":

```
FromIO(decode)
  .FlatMapIO(validate)          // Left(VALIDATION_FAILED) breaks the branch
  .FlatMapIO(repo.Insert)       // Left(INTERNAL) breaks the branch
  .MapIO(toResponse)            // executed only on Right
  .Run(ctx)                     // single effect execution point
```

---

## 7. Context and Transaction Management

- `Run(ctx)` receives the request `context.Context`; cancellation/timeout is passed to the DB.
- Transactions are encapsulated in the postgres adapter: `Begin → steps → Commit/Rollback`
  (see REQ-DB-004). The service layer does not know `Tx` details.
- Long operations are limited by `context.WithTimeout`.

---

## 8. Dependencies (Stack)

| Area | Choice | Rationale |
|---|---|---|
| Language | Go | requirement |
| HTTP | `net/http` + light router | no heavy frameworks, fewer dependencies |
| DB | PostgreSQL + `database/sql`/`pgx` | requirement |
| Migrations | go-migrations (e.g. `golang-migrate`) or SQL files | versioned schema |
| Auth | JWT (HS256) + Argon2id | requirement |
| Frontend | TypeScript | requirement |
| Environment | Docker Compose + `.env` | requirement |

---

## 9. Go Generic Limitations Affecting Design

1. Methods cannot introduce their own type parameters → `Map`/`FlatMap`
   are implemented as free generic functions.
2. `Either[E,T]` is more convenient as a struct with a discriminator than as an interface
   (otherwise dynamic dispatch and loss of type safety).
3. `IOEither` can be made an interface with generic methods `MapIO[U]`, `FlatMapIO[U]`,
   but `Run` returns a concrete `Either[E,T]` — type parameters are fixed
   at creation time.
4. Type information does not "stretch" automatically across long chains —
   composition is built explicitly, improving ROP pipeline readability.

---

## 10. Project Catalog (suggested layout)

```
go/
├── cmd/server/main.go            entry point
├── internal/
│   ├── config/                   environment configuration
│   ├── domain/                   User, Guest, DomainError, factories
│   ├── rop/                      Either, IOEither
│   ├── service/                  use-cases (GuestService, AuthService)
│   ├── repository/               interfaces (port)
│   │   └── postgres/             adapters (T, error)
│   └── transport/
│       ├── http/                 handlers, routes
│       └── middleware/           JWT
├── migrations/                   SQL migrations
├── web/                          frontend TypeScript
├── docker-compose.yml
├── .env.example
└── README.md
```
