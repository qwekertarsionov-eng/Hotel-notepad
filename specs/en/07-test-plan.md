# 07 — Test Plan (Test Plan)

**System:** Guest Accounting
**Language:** English
**Tools:** Go `testing`, `httptest`, integration tests with PostgreSQL,
TypeScript build for FE.

Test ID format: `TST-XXX-NNN` (XXX — domain, NNN — number).
Matrix format: `Requirement ID → Test ID → Expected Behavior → Test File`.

---

## 1. Test ID Domain Notation

| Domain | Prefix | Test type |
|---|---|---|
| ROP abstractions | `TST-ROP-*` | unit |
| Authentication | `TST-AUTH-*` | HTTP (httptest) |
| Guests/CRUD/isolation | `TST-GUEST-*` | HTTP (httptest) + integration |
| API contract | `TST-API-*` | HTTP (httptest) |
| Database | `TST-DB-*` | integration (PostgreSQL) |
| Security | `TST-SEC-*` | unit + HTTP |
| Frontend | `TST-FE-*` | build/types |
| Infrastructure | `TST-INFRA-*` | integration/build |

Suggested test files:
- `internal/rop/either_test.go`, `internal/rop/ioeither_test.go`
- `internal/transport/http/auth_test.go`
- `internal/transport/http/guest_test.go`
- `internal/transport/http/api_test.go`
- `internal/repository/postgres/*_integration_test.go`
- `internal/domain/guest_validation_test.go`
- `internal/service/*_test.go`
- `web/*.test.ts` / build check `tsc`

---

## 2. Matrix REQ → Test ID → Expected Behavior → Test File

### 2.1 ROP abstractions (`TST-ROP-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-ROP-001,003 | `TST-ROP-001` | `Either` cannot be simultaneously `Left` and `Right`; always exactly one state | `internal/rop/either_test.go` |
| REQ-ROP-001,003 | `TST-ROP-002` | `FlatMap(Left, f)` does not call `f` (counter = 0), returns original `Left` | `internal/rop/either_test.go` |
| REQ-ROP-001,003 | `TST-ROP-003` | `Map(Left, f)` preserves original error, `f` is not called | `internal/rop/either_test.go` |
| REQ-ROP-002,003 | `TST-ROP-004` | until `Run` the `IOEither` effect is not executed (counter = 0) | `internal/rop/ioeither_test.go` |
| REQ-ROP-002,003 | `TST-ROP-005` | `FromIO(f)` on `error` returns `Left` with same cause | `internal/rop/ioeither_test.go` |
| REQ-ROP-002,003 | `TST-ROP-006` | cancelled `ctx` is passed into effect → `Run` returns `Left` | `internal/rop/ioeither_test.go` |
| REQ-ROP-002,003 | `TST-ROP-007` | two `Run`s execute effect twice (effect is not memoized) | `internal/rop/ioeither_test.go` |
| REQ-ROP-001 | `TST-ROP-008` | `Fold` folds `Left` through first function, `Right` — through second | `internal/rop/either_test.go` |
| REQ-ROP-001 | `TST-ROP-009` | `MapLeft` transforms error, `Right` passes through | `internal/rop/either_test.go` |
| REQ-ROP-001 | `TST-ROP-010` | `FlatMap` is associative (monad law) | `internal/rop/either_test.go` |
| REQ-ROP-001 | `TST-ROP-011` | `Map` through `FlatMap` is consistent with direct `Map` | `internal/rop/either_test.go` |
| REQ-ROP-002 | `TST-ROP-012` | `MapIO`/`FlatMapIO` preserve laziness until `Run` | `internal/rop/ioeither_test.go` |
| REQ-ROP-002 | `TST-ROP-013` | `FromIO` of a successful value → `Right` | `internal/rop/ioeither_test.go` |
| REQ-ROP-002 | `TST-ROP-014` | `FromIO` error/panic → `Left` with cause (no loss) | `internal/rop/ioeither_test.go` |

### 2.2 Authentication (`TST-AUTH-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-AUTH-001 | `TST-AUTH-001` | successful registration → `201`, 1 `users` record created | `internal/transport/http/auth_test.go` |
| REQ-AUTH-001 | `TST-AUTH-002` | repeat registration of existing email → `409` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-001 | `TST-AUTH-003` | registration with invalid email/short password → `422` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-002 | `TST-AUTH-004` | correct credentials → `200` + valid JWT | `internal/transport/http/auth_test.go` |
| REQ-AUTH-002 | `TST-AUTH-005` | login with wrong password → `401` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-002 | `TST-AUTH-006` | login with nonexistent email → `401` (same message as wrong password) | `internal/transport/http/auth_test.go` |
| REQ-AUTH-003 | `TST-AUTH-007` | request without token to `/guests` → `401` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-003 | `TST-AUTH-008` | request with invalid signature → `401` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-003 | `TST-AUTH-009` | request with expired token → `401` | `internal/transport/http/auth_test.go` |
| REQ-AUTH-001, SEC-001 | `TST-AUTH-010` | password stored in DB as hash, not plaintext | `internal/transport/http/auth_test.go` |

### 2.3 Guests: CRUD, search, isolation (`TST-GUEST-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-GUEST-001 | `TST-GUEST-001` | guest creation → `201`, record returned with `id` and owner `user_id` | `internal/transport/http/guest_test.go` |
| REQ-API-003, GUEST-001 | `TST-GUEST-002` | creation with empty full name → `422`, field `full_name` in `fields` | `internal/transport/http/guest_test.go` |
| REQ-API-003, GUEST-001 | `TST-GUEST-003` | creation with `check_out <= check_in` → `422` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-001 | `TST-GUEST-004` | creation without token → `401` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-002 | `TST-GUEST-005` | list → `200`, correct `items/page/limit/total`, only own records | `internal/transport/http/guest_test.go` |
| REQ-GUEST-003 | `TST-GUEST-006` | search by full name substring finds record → `200` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-003 | `TST-GUEST-007` | search without results → `200` + empty `items` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-004 | `TST-GUEST-008` | view own record → `200` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-004,007 | `TST-GUEST-009` | view someone else's record → `404` (indistinguishable from nonexistent) | `internal/transport/http/guest_test.go` |
| REQ-GUEST-005 | `TST-GUEST-010` | edit own record → `200`, fields updated | `internal/transport/http/guest_test.go` |
| REQ-GUEST-005,007 | `TST-GUEST-011` | edit someone else's record → `404`, data unchanged | `internal/transport/http/guest_test.go` |
| REQ-GUEST-006 | `TST-GUEST-012` | delete own record → `204`, repeat GET → `404` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-006,007 | `TST-GUEST-013` | delete someone else's record → `404`, record exists | `internal/transport/http/guest_test.go` |
| REQ-GUEST-004 | `TST-GUEST-014` | view nonexistent record → `404` | `internal/transport/http/guest_test.go` |
| REQ-GUEST-002,007 | `TST-GUEST-015` | isolation: user A's list/search does not contain B's records | `internal/transport/http/guest_test.go` |

### 2.4 API contract (`TST-API-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-API-001 | `TST-API-001` | matrix "endpoint → expected status" is fulfilled | `internal/transport/http/api_test.go` |
| REQ-API-002 | `TST-API-002` | all errors contain `code` + safe `message` | `internal/transport/http/api_test.go` |
| REQ-API-003 | `TST-API-003` | invalid JSON in body → `400` | `internal/transport/http/api_test.go` |
| REQ-API-001 | `TST-API-004` | unknown path → `404` | `internal/transport/http/api_test.go` |
| REQ-API-001 | `TST-API-005` | unsupported method → `405` | `internal/transport/http/api_test.go` |

### 2.5 Database (`TST-DB-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-DB-001 | `TST-DB-001` | migrations applied to clean DB without errors | `internal/repository/postgres/migrations_integration_test.go` |
| REQ-DB-001 | `TST-DB-002` | re-application of applied migrations does not break schema | `internal/repository/postgres/migrations_integration_test.go` |
| REQ-DB-002 | `TST-DB-003` | inserting guest with nonexistent `user_id` rejected (FK) | `internal/repository/postgres/guests_integration_test.go` |
| REQ-DB-002 | `TST-DB-004` | `NOT NULL` violation rejected | `internal/repository/postgres/guests_integration_test.go` |
| REQ-DB-002 | `TST-DB-005` | record with `check_out <= check_in` rejected (CHECK) | `internal/repository/postgres/guests_integration_test.go` |
| REQ-DB-003, GUEST-003 | `TST-DB-006` | `EXPLAIN ANALYZE` of list query uses index `(user_id, created_at)` | `internal/repository/postgres/guests_integration_test.go` |
| REQ-DB-004 | `TST-DB-007` | failure mid-transaction leaves no partial records (ROLLBACK) | `internal/repository/postgres/tx_integration_test.go` |

### 2.6 Security (`TST-SEC-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-SEC-001, AUTH-001 | `TST-SEC-001` | plaintext password not stored in DB (only hash) | `internal/transport/http/auth_test.go` |
| REQ-SEC-002, AUTH-002 | `TST-SEC-002` | JWT signature verified; forged token → `401` | `internal/transport/http/auth_test.go` |
| REQ-SEC-002, AUTH-003 | `TST-SEC-003` | responses do not contain `JWT_SECRET` value | `internal/transport/http/api_test.go` |
| REQ-SEC-003 | `TST-SEC-004` | `500` body contains no SQL/stack/paths/tables | `internal/transport/http/api_test.go` |

### 2.7 Frontend (`TST-FE-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-FE-001 | `TST-FE-001` | TypeScript project compiles with `tsc` without errors | `web/tsconfig` + CI build step |
| REQ-FE-002 | `TST-FE-002` | protected route without token redirects to login | `web/router.test.ts` |
| REQ-FE-002 | `TST-FE-003` | after login token is saved and added to `Authorization` | `web/auth.test.ts` |

### 2.8 Infrastructure (`TST-INFRA-*`)

| REQ | Test ID | Expected Behavior | Test File |
|---|---|---|---|
| REQ-INFRA-001 | `TST-INFRA-001` | `docker compose up` deploys stack; endpoint responds | manual check / script `scripts/smoke.sh` |
| REQ-INFRA-002 | `TST-INFRA-002` | `.env.example` contains all application-read variables | `internal/config/config_test.go` |
| REQ-INFRA-003 | `TST-INFRA-003` | README commands execute and work | code review / script |
| REQ-INFRA-004 | `TST-INFRA-004` | `go test ./...` passes (unit + HTTP) | CI |
| REQ-INFRA-004 | `TST-INFRA-005` | integration tests with PostgreSQL pass when DB available | CI (with DB service) |

---

## 3. Requirements Coverage

| REQ domain | Count of REQ | Covered by Test ID |
|---|---|---|
| REQ-AUTH-* | 3 | `TST-AUTH-001..010`, `TST-SEC-001,002` |
| REQ-GUEST-* | 7 | `TST-GUEST-001..015` |
| REQ-ROP-* | 3 | `TST-ROP-001..014` |
| REQ-DB-* | 4 | `TST-DB-001..007` |
| REQ-SEC-* | 3 | `TST-SEC-001..004` |
| REQ-API-* | 3 | `TST-API-001..005` |
| REQ-FE-* | 2 | `TST-FE-001..003` |
| REQ-INFRA-* | 4 | `TST-INFRA-001..005` |

Each requirement has ≥ 1 related Test ID (traceability complete).

---

## 4. Test Types and How to Run Them

| Type | Command | Dependencies |
|---|---|---|
| Unit (ROP, domain) | `go test ./internal/rop/... ./internal/domain/...` | none |
| HTTP (httptest) | `go test ./internal/transport/...` | repository mocks |
| Integration (DB) | `go test -tags=integration ./internal/repository/...` | PostgreSQL (Docker) |
| Go build | `go build ./...` | none |
| FE build | `npm run build` (tsc) | Node |

Integration tests must:
- use an isolated DB/schema (e.g. separate DB per test run);
- skip (`t.Skip`) when DB unavailable and no tag/variable present;
- clean data deterministically.

---

## 5. Special Check Categories

| Category | Test ID | Essence |
|---|---|---|
| Monad laws | `TST-ROP-010`, `TST-ROP-011` | `FlatMap` associativity, `Map` consistency |
| Effect laziness | `TST-ROP-004`, `TST-ROP-012` | no side effects before `Run` |
| Context cancellation | `TST-ROP-006` | `ctx` propagated to effects |
| Data isolation | `TST-GUEST-009,011,013,015` | someone else's records inaccessible |
| Detail leakage | `TST-SEC-004` | `500` without SQL/stack |
| PostgreSQL constraints | `TST-DB-003..007` | FK/NOT NULL/CHECK/index/transactions |
