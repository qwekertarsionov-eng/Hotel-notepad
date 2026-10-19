# 08 — Acceptance Criteria (Acceptance Criteria)

**System:** Guest Accounting
**Language:** English

This section fixes verifiable completion conditions per requirement and common
(GATE) release criteria. Traceability to tests — in `07-test-plan.md`.

Notation: a criterion is considered accepted when related Test IDs are green.

---

## 1. Criteria per Requirement

### REQ-AUTH-001 — Registration
- **AC-AUTH-001a:** `POST /auth/register` with valid data creates 1 `users` record and returns `201`. (`TST-AUTH-001`)
- **AC-AUTH-001b:** repeat registration of the same email → `409`. (`TST-AUTH-002`)
- **AC-AUTH-001c:** invalid email / short password → `422` with field list. (`TST-AUTH-003`)
- **AC-AUTH-001d:** DB stores only password hash. (`TST-AUTH-010`, `TST-SEC-001`)

### REQ-AUTH-002 — Login and Token
- **AC-AUTH-002a:** correct credentials → `200` + valid JWT. (`TST-AUTH-004`)
- **AC-AUTH-002b:** wrong password → `401`. (`TST-AUTH-005`)
- **AC-AUTH-002c:** nonexistent email → `401`, message indistinguishable from wrong password. (`TST-AUTH-006`)

### REQ-AUTH-003 — JWT middleware
- **AC-AUTH-003a:** request without token → `401`. (`TST-AUTH-007`)
- **AC-AUTH-003b:** forged signature → `401`. (`TST-AUTH-008`)
- **AC-AUTH-003c:** expired token → `401`. (`TST-AUTH-009`)

### REQ-GUEST-001 — Guest creation
- **AC-GUEST-001a:** valid body → `201`, `user_id` = token owner. (`TST-GUEST-001`)
- **AC-GUEST-001b:** `user_id` is not taken from request body. (`TST-GUEST-001`)

### REQ-GUEST-002 — Guest list
- **AC-GUEST-002a:** list contains only owner records. (`TST-GUEST-005`, `TST-GUEST-015`)
- **AC-GUEST-002b:** pagination returns correct `page/limit/total`. (`TST-GUEST-005`)

### REQ-GUEST-003 — Search
- **AC-GUEST-003a:** search by full name substring finds owner's record. (`TST-GUEST-006`)
- **AC-GUEST-003b:** empty result → `200` + empty `items`. (`TST-GUEST-007`)

### REQ-GUEST-004 — View
- **AC-GUEST-004a:** own record → `200`. (`TST-GUEST-008`)
- **AC-GUEST-004b:** someone else's and nonexistent record → identical `404`. (`TST-GUEST-009`, `TST-GUEST-014`)

### REQ-GUEST-005 — Editing
- **AC-GUEST-005a:** updating own record → `200`, only passed fields changed. (`TST-GUEST-010`)
- **AC-GUEST-005b:** someone else's record → `404`, data unchanged. (`TST-GUEST-011`)

### REQ-GUEST-006 — Deletion
- **AC-GUEST-006a:** deleting own record → `204`; repeat GET → `404`. (`TST-GUEST-012`)
- **AC-GUEST-006b:** someone else's record is not deleted → `404`. (`TST-GUEST-013`)

### REQ-GUEST-007 — Isolation
- **AC-GUEST-007a:** user A does not view/edit/delete records of B. (`TST-GUEST-009,011,013,015`)

### REQ-ROP-001 — Either
- **AC-ROP-001a:** `Map`/`FlatMap`/`MapLeft`/`Fold` behave per contract (including `FlatMap(Left)` does not call function, `Map(Left)` preserves error). (`TST-ROP-002,003,008,009`)

### REQ-ROP-002 — IOEither
- **AC-ROP-002a:** effect does not execute before `Run`. (`TST-ROP-004`, `TST-ROP-012`)
- **AC-ROP-002b:** `Run(ctx)` executes effect and respects context cancellation. (`TST-ROP-006`)

### REQ-ROP-003 — 7 ROP invariants
- **AC-ROP-003a:** all 7 invariants covered by separate tests and green. (`TST-ROP-001..007`)

### REQ-DB-001 — Migrations
- **AC-DB-001a:** migrations applied to clean DB. (`TST-DB-001`)
- **AC-DB-001b:** repeat application does not break schema. (`TST-DB-002`)

### REQ-DB-002 — Integrity
- **AC-DB-002a:** nonexistent `user_id` rejected (FK). (`TST-DB-003`)
- **AC-DB-002b:** `NOT NULL` violation → rejection. (`TST-DB-004`)
- **AC-DB-002c:** `check_out <= check_in` rejected at DB level. (`TST-DB-005`)

### REQ-DB-003 — Index
- **AC-DB-003a:** `EXPLAIN ANALYZE` of typical query uses `(user_id, created_at)`. (`TST-DB-006`)

### REQ-DB-004 — Transactions
- **AC-DB-004a:** mid-transaction failure leaves no partial data. (`TST-DB-007`)

### REQ-SEC-001 — Password hashing
- **AC-SEC-001a:** plaintext password absent in DB and logs. (`TST-SEC-001`)

### REQ-SEC-002 — JWT/secret
- **AC-SEC-002a:** forged token rejected. (`TST-SEC-002`)
- **AC-SEC-002b:** responses do not contain secret. (`TST-SEC-003`)

### REQ-SEC-003 — No detail leakage
- **AC-SEC-003a:** `500` body contains no SQL/stack/paths/tables. (`TST-SEC-004`)

### REQ-API-001 — Endpoints/statuses
- **AC-API-001a:** endpoint→status matrix fully fulfilled. (`TST-API-001`)
- **AC-API-001b:** unknown path → `404`, unsupported method → `405`. (`TST-API-004`, `TST-API-005`)

### REQ-API-002 — Error format
- **AC-API-002a:** each error contains `code` + safe `message`. (`TST-API-002`)

### REQ-API-003 — Boundary validation
- **AC-API-003a:** invalid JSON → `400`; semantically invalid values → `422`. (`TST-API-003`, `TST-GUEST-002,003`)

### REQ-FE-001 — Frontend
- **AC-FE-001a:** TypeScript project compiles without errors. (`TST-FE-001`)

### REQ-FE-002 — Token/routes
- **AC-FE-002a:** protected route without token redirects to login. (`TST-FE-002`)
- **AC-FE-002b:** after login token added to `Authorization`. (`TST-FE-003`)

### REQ-INFRA-001 — Docker Compose
- **AC-INFRA-001a:** `docker compose up` deploys working stack. (`TST-INFRA-001`)

### REQ-INFRA-002 — .env.example
- **AC-INFRA-002a:** `.env.example` contains all application-read variables; launch from template works. (`TST-INFRA-002`)

### REQ-INFRA-003 — README
- **AC-INFRA-003a:** README commands (stack, launch, tests, migrations, env) executable. (`TST-INFRA-003`)

### REQ-INFRA-004 — Test infrastructure
- **AC-INFRA-004a:** `go test ./...` passes. (`TST-INFRA-004`)
- **AC-INFRA-004b:** integration tests with PostgreSQL pass when DB available. (`TST-INFRA-005`)

---

## 2. Common GATE Release Criteria

GATE — a gateway that must be fully green before system acceptance.

| ID | GATE criterion | Check |
|---|---|---|
| **GATE-1** | All REQ have ≥ 1 related Test ID, and no Test ID "hangs" without REQ | traceability 01 ↔ 07 |
| **GATE-2** | `go build ./...` and `go vet ./...` without errors | CI |
| **GATE-3** | `go test ./...` fully green | CI |
| **GATE-4** | Integration tests with PostgreSQL green when DB available | CI |
| **GATE-5** | TypeScript build (`tsc`) without errors | CI |
| **GATE-6** | 7 ROP invariants confirmed by tests (`TST-ROP-001..007`) | tests |
| **GATE-7** | Full user isolation confirmed (`TST-GUEST-009,011,013,015`) | tests |
| **GATE-8** | No response reveals SQL/stack/secret (`TST-SEC-003,004`) | tests |
| **GATE-9** | Migrations applied from zero and idempotent (`TST-DB-001,002`) | tests |
| **GATE-10** | `.env.example` + README allow a newcomer to launch the project | manual check |
| **GATE-11** | `docker compose up` gives working stack | smoke test |
| **GATE-12** | Index `(user_id, created_at)` confirmed by `EXPLAIN ANALYZE` | `TST-DB-006` |

### Readiness Threshold

The system is considered ready for acceptance when:
1. all AC per each REQ are fulfilled (section 1);
2. all GATE-1…GATE-12 are fulfilled;
3. there are no open "blocker"/"critical" defects;
4. traceability REQ → Test → AC is non-contradictory.

---

## 3. Non-functional Acceptance Conditions

| Area | Condition |
|---|---|
| Type safety | no `any`/unsafe casts in domain and ROP layers |
| Password security | only Argon2id/bcrypt; no plaintext passwords |
| Configuration | secrets only from environment, no hardcoded values |
| Observability | errors logged on server (with cause), client gets safe message |
| Reproducibility | launch and tests reproducible via README/Compose |
