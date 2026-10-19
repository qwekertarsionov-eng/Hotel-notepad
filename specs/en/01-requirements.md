# 01 — System Requirements (Requirements)

**System:** Guest Accounting (hotel accounting)
**Version:** 1.0
**Status:** Approved
**Language:** English

Each requirement is described by eight fields:

1. **Description** — what should be implemented.
2. **Preconditions** — system state before execution.
3. **Input data** — what is submitted at the boundary.
4. **Expected behavior** — observable result.
5. **Errors** — failure conditions and their mapping.
6. **Invariants** — conditions that always hold.
7. **Acceptance criteria** — verifiable completion condition.
8. **Related Test IDs** — traceability to test plan (07).

Error code designations are given in `06-error-model.md`.
Response format is in `04-api-contract.md`.

---

## Requirements by domain

| Domain | Prefix | Count |
|---|---|---|
| Authentication | `REQ-AUTH-*` | 3 |
| Guests | `REQ-GUEST-*` | 7 |
| ROP abstractions | `REQ-ROP-*` | 3 |
| Database | `REQ-DB-*` | 4 |
| Security | `REQ-SEC-*` | 3 |
| API contract | `REQ-API-*` | 3 |
| Frontend | `REQ-FE-*` | 2 |
| Infrastructure | `REQ-INFRA-*` | 4 |
| **Total** | | **29** |

---

# REQ-AUTH-001 — User Registration

- **Description:** The system allows creating a user account by email and password. The password is stored only as a hash.
- **Preconditions:** A user with the given email does not yet exist. The `users` table is writable.
- **Input data:** JSON body `POST /api/v1/auth/register`: `email` (string, valid format), `password` (string, minimum 8 characters).
- **Expected behavior:** A `users` record is created with a unique email and a password hash (Argon2id). Returns `201 Created` with confirmation (id, email) without the hash and without the plaintext password.
- **Errors:** email already taken → `409 CONFLICT` (`AUTH_EMAIL_TAKEN`); invalid data → `422 UNPROCESSABLE` (`VALIDATION_FAILED`); internal DB error → `500 INTERNAL`.
- **Invariants:** plaintext password is never stored and never returned; `email` is unique; `created_at` is set automatically.
- **Acceptance criteria:** successful registration creates one row in `users`; repeat registration of the same email returns `409`; hash is not equal to the plaintext password.
- **Related Test IDs:** `TST-AUTH-001`, `TST-AUTH-002`, `TST-AUTH-003`, `TST-AUTH-010`, `TST-SEC-001`.

---

# REQ-AUTH-002 — Authentication and JWT Token Issuance

- **Description:** The system authenticates a user by email and password and issues a JWT access token.
- **Preconditions:** User account exists (see REQ-AUTH-001).
- **Input data:** JSON body `POST /api/v1/auth/login`: `email`, `password`.
- **Expected behavior:** Password hash is compared with the stored one (constant-time comparison). On success returns `200 OK` and a JWT token (`Authorization` header and/or response body) with a lifetime.
- **Errors:** wrong password or nonexistent email → `401 UNAUTHORIZED` (`AUTH_INVALID_CREDENTIALS`) with the same message for both cases; invalid data → `422 UNPROCESSABLE` (`VALIDATION_FAILED`).
- **Invariants:** response does not reveal whether the email exists; password is not returned; token is signed with the server secret.
- **Acceptance criteria:** correct credentials return a valid JWT; incorrect ones return `401`; response time does not depend on email existence.
- **Related Test IDs:** `TST-AUTH-004`, `TST-AUTH-005`, `TST-AUTH-006`, `TST-SEC-002`.

---

# REQ-AUTH-003 — API Protection via JWT Middleware

- **Description:** All `guests` resource endpoints require a valid JWT in the `Authorization: Bearer <token>` header.
- **Preconditions:** Token obtained (REQ-AUTH-002). Server knows the signing secret.
- **Input data:** HTTP `Authorization` header with Bearer token; request path and method.
- **Expected behavior:** Middleware verifies signature, lifetime, and extracts `user_id`. On success the request is passed further with an authenticated context; on failure it is not passed.
- **Errors:** missing header → `401 UNAUTHORIZED` (`AUTH_TOKEN_MISSING`); invalid signature → `401` (`AUTH_TOKEN_INVALID`); expired lifetime → `401` (`AUTH_TOKEN_EXPIRED`).
- **Invariants:** unauthenticated request never reaches the service layer; forged token is rejected; `user_id` is taken from the token, not from the request body.
- **Acceptance criteria:** request without/with broken/with expired token returns `401`; request with valid token passes.
- **Related Test IDs:** `TST-AUTH-007`, `TST-AUTH-008`, `TST-AUTH-009`, `TST-SEC-002`.

---

# REQ-GUEST-001 — Guest Record Creation

- **Description:** Authenticated user creates a guest record with fields: full name, phone, room number, check-in date, check-out date, note.
- **Preconditions:** User authenticated (REQ-AUTH-003). Data passed validation (REQ-GUEST domain + REQ-API-003).
- **Input data:** JSON body `POST /api/v1/guests`: `full_name`, `phone`, `room_number`, `check_in`, `check_out`, `note` (optional). `user_id` is taken from the token.
- **Expected behavior:** A row is created in `guests` with the owner's `user_id`. Returns `201 Created` and the full record including `id`.
- **Errors:** invalid fields (empty full name, `check_out <= check_in`) → `422 UNPROCESSABLE` (`VALIDATION_FAILED`); no token → `401`; internal error → `500 INTERNAL`.
- **Invariants:** `user_id` always equals the authenticated user and is not taken from the body; record belongs to exactly one owner.
- **Acceptance criteria:** valid body creates a record and returns `201`; `user_id` in DB matches the token owner.
- **Related Test IDs:** `TST-GUEST-001`, `TST-GUEST-002`, `TST-GUEST-003`, `TST-GUEST-004`.

---

# REQ-GUEST-002 — Guest List Retrieval

- **Description:** Authenticated user gets a paginated list of guests belonging to them.
- **Preconditions:** User authenticated. The `guests` table may contain records of different owners.
- **Input data:** Query parameters `GET /api/v1/guests`: `page`, `limit` (optional, default values).
- **Expected behavior:** Returns `200 OK` and an array of the current user's records + pagination metadata. Other users' records are not included.
- **Errors:** no token → `401`; incorrect pagination parameters → `400 BAD_REQUEST`.
- **Invariants:** selection is always filtered by the owner's `user_id`; order is deterministic (e.g., by `created_at DESC`).
- **Acceptance criteria:** user sees only their records; pagination returns correct `page/limit/total`.
- **Related Test IDs:** `TST-GUEST-005`, `TST-GUEST-015`.

---

# REQ-GUEST-003 — Guest Search

- **Description:** Authenticated user searches their guests by full name substring, phone, or room number.
- **Preconditions:** User authenticated.
- **Input data:** Query parameter `GET /api/v1/guests/search`: `q` (non-empty string); optional `field` (search field).
- **Expected behavior:** Returns `200 OK` and an array of matching records of the current user. Search is case-insensitive by full name.
- **Errors:** empty `q` → `400 BAD_REQUEST`; no token → `401`.
- **Invariants:** search does not go beyond the owner's records; empty result is not an error (`200` + empty array).
- **Acceptance criteria:** full name substring match finds a record; search returns only owner's records.
- **Related Test IDs:** `TST-GUEST-006`, `TST-GUEST-007`, `TST-DB-006`.

---

# REQ-GUEST-004 — Viewing a Single Guest Record

- **Description:** Authenticated user gets a single guest record by `id`.
- **Preconditions:** User authenticated.
- **Input data:** Path parameter `GET /api/v1/guests/{id}`.
- **Expected behavior:** If the record exists and belongs to the user — `200 OK` and the record. If the record exists but belongs to another user — `404` (existence of another user's record is not revealed).
- **Errors:** record not found or belongs to another → `404 NOT_FOUND` (`GUEST_NOT_FOUND`); no token → `401`.
- **Invariants:** another user's record is indistinguishable from a nonexistent one (single `404`).
- **Acceptance criteria:** owner gets `200`; stranger and nonexistent id — identical `404`.
- **Related Test IDs:** `TST-GUEST-008`, `TST-GUEST-009`, `TST-GUEST-014`.

---

# REQ-GUEST-005 — Editing a Guest Record

- **Description:** Authenticated user updates fields of their guest record.
- **Preconditions:** User authenticated; record belongs to the user.
- **Input data:** Path parameter `{id}` + JSON body `PUT /api/v1/guests/{id}` with fields to update.
- **Expected behavior:** Fields are validated and updated. Returns `200 OK` and the updated record.
- **Errors:** someone else's/nonexistent record → `404` (`GUEST_NOT_FOUND`); invalid fields (including `check_out <= check_in`) → `422`; no token → `401`.
- **Invariants:** `user_id` and `id` are not changed via request body; date invariant is preserved after update.
- **Acceptance criteria:** update changes only passed fields; someone else's record is inaccessible (`404`).
- **Related Test IDs:** `TST-GUEST-010`, `TST-GUEST-011`.

---

# REQ-GUEST-006 — Deleting a Guest Record

- **Description:** Authenticated user deletes their guest record.
- **Preconditions:** User authenticated; record belongs to the user.
- **Input data:** Path parameter `DELETE /api/v1/guests/{id}`.
- **Expected behavior:** Record is deleted. Returns `204 No Content` without body.
- **Errors:** someone else's/nonexistent record → `404` (`GUEST_NOT_FOUND`); no token → `401`.
- **Invariants:** deletion affects only the owner's row; after deletion the record is unavailable via GET as well.
- **Acceptance criteria:** `DELETE` returns `204`; repeat GET returns `404`; someone else's record is not deleted (`404`).
- **Related Test IDs:** `TST-GUEST-012`, `TST-GUEST-013`.

---

# REQ-GUEST-007 — User Data Isolation

- **Description:** Data of different users is strictly isolated: a user cannot read, modify, or delete someone else's records via any endpoint.
- **Preconditions:** System has two or more users with their own records.
- **Input data:** any request to `/api/v1/guests*` with user A's token affecting user B's record.
- **Expected behavior:** User A does not receive, change, or delete records of B. Operations on someone else's records return `404`.
- **Errors:** attempt to access someone else's record → `404` (`GUEST_NOT_FOUND`).
- **Invariants:** each SQL query of the repository layer is filtered by `user_id` from the token; filtering is not optional.
- **Acceptance criteria:** A's list/search does not contain B's records; direct GET/PUT/DELETE by someone else's id returns `404`; B's data unchanged.
- **Related Test IDs:** `TST-GUEST-009`, `TST-GUEST-011`, `TST-GUEST-013`, `TST-GUEST-015`.

---

# REQ-ROP-001 — Either[E,T] Abstraction

- **Description:** In the business layer the result of an operation is expressed as type `Either[E,T]` with exactly one state: `Left[E]` (error) or `Right[T]` (success). Methods `Map`, `FlatMap`, `MapLeft`, `Fold` are implemented.
- **Preconditions:** Error domain `E` is defined (see 06-error-model.md).
- **Input data:** original value `Either[E,T]` and transformation functions.
- **Expected behavior:** `Map` applies the function to `Right` and ignores `Left`; `MapLeft` applies the function to `Left` and ignores `Right`; `FlatMap` passes `Right` to the next function returning `Either`; `Fold` folds both states into one value.
- **Errors:** data errors are `Left[E]`, not Go exceptions.
- **Invariants:** invariants 1–3 from REQ-ROP-003 are observed.
- **Acceptance criteria:** Either laws are verified by tests (see 07-test-plan.md).
- **Related Test IDs:** `TST-ROP-001`, `TST-ROP-002`, `TST-ROP-003`, `TST-ROP-008`, `TST-ROP-009`, `TST-ROP-010`, `TST-ROP-011`.

---

# REQ-ROP-002 — IOEither[E,T] Abstraction

- **Description:** Effects (DB accesses, network) are encapsulated in the lazy type `IOEither[E,T]`, which does no work until `Run(ctx)` is called. `FromIO`, `MapIO`, `FlatMapIO`, `Run(ctx)` are implemented.
- **Preconditions:** There is `Either[E,T]` (REQ-ROP-001) and `context.Context`.
- **Input data:** effect function returning `Either[E,T]` (or `IOEither`); context for `Run`.
- **Expected behavior:** constructor `FromIO` wraps the effect without executing it; `MapIO`/`FlatMapIO` build a new description; `Run(ctx)` executes the effect and returns `Either[E,T]`, respecting context cancellation.
- **Errors:** effect failure becomes `Left[E]`; context cancellation is also `Left[E]` (context error).
- **Invariants:** invariants 4–7 from REQ-ROP-003 are observed.
- **Acceptance criteria:** before `Run` there are no side effects; `Run` executes the effect once; context cancellation interrupts execution.
- **Related Test IDs:** `TST-ROP-004`, `TST-ROP-005`, `TST-ROP-006`, `TST-ROP-007`, `TST-ROP-012`, `TST-ROP-013`, `TST-ROP-014`.

---

# REQ-ROP-003 — Seven ROP Invariants Observed

- **Description:** `Either`/`IOEither` implementation satisfies the seven Railway-Oriented Programming invariants.
- **Preconditions:** Types implemented (REQ-ROP-001, REQ-ROP-002).
- **Input data:** arbitrary `Either`/`IOEither` values and functions.
- **Expected behavior:** All seven invariants are confirmed by autotests (see 07-test-plan.md, domain `TST-ROP-*`):
  1. `Either` contains exactly one valid state (not `Left` and `Right` simultaneously, not "empty").
  2. `FlatMap(Left)` does not call the next function.
  3. `Map(Left)` preserves the original error.
  4. `IOEither` does not execute the effect before `Run`.
  5. `FromIO` does not lose the error cause (effect error becomes `Left` with the same cause).
  6. Context cancellation is passed to nested effects.
  7. Repeated `Run` may repeat the effect (effect is not memoized without an explicit cache).
- **Errors:** violation of any invariant is an implementation defect, test fails.
- **Invariants:** the listed seven conditions are true for all inputs.
- **Acceptance criteria:** each of the seven invariants is covered by a separate test, all green.
- **Related Test IDs:** `TST-ROP-001` … `TST-ROP-007`.

Note on applicability boundary: the `(T, error)` type is standard for infrastructure interfaces (repository, HTTP clients). `IOEither` is a composition method, not a reliability guarantee (does not cancel the need for retries, timeouts, transactions).

---

# REQ-DB-001 — DB Schema Migrations

- **Description:** DB schema (`users`, `guests`) is managed by versioned migrations applicable from a clean state.
- **Preconditions:** Empty (or previous version) PostgreSQL is available.
- **Input data:** set of migration files (see 05-database.md) + migration application command.
- **Expected behavior:** Migrations are applied in version order, creating tables, constraints, and indexes. Re-application of already applied migrations does not break the schema.
- **Errors:** migration syntax error → migration is rolled back, command completes with error.
- **Invariants:** migration order is deterministic; each migration is idempotent on application (tracked in a service table).
- **Acceptance criteria:** on a clean DB all migrations are applied; repeat run creates no duplicates and does not fail.
- **Related Test IDs:** `TST-DB-001`, `TST-DB-002`.

---

# REQ-DB-002 — Data Integrity (FK, NOT NULL, CHECK)

- **Description:** Referential integrity and validity constraints are ensured at the DB level.
- **Preconditions:** Schema applied (REQ-DB-001).
- **Input data:** INSERT/UPDATE operations violating constraints.
- **Expected behavior:** `guests.user_id` — foreign key to `users.id`; mandatory fields — `NOT NULL`; date interval verified by `CHECK (check_out > check_in)`.
- **Errors:** constraint violation → PostgreSQL error, convertible to `422`/`500` without leaking SQL details.
- **Invariants:** nonexistent `user_id` cannot be written; dates are always consistent at DB level (insurance above application validation).
- **Acceptance criteria:** inserting a guest with nonexistent `user_id` is rejected; record with `check_out <= check_in` is rejected even bypassing API.
- **Related Test IDs:** `TST-DB-003`, `TST-DB-004`, `TST-DB-005`.

---

# REQ-DB-003 — Grounded PostgreSQL Index

- **Description:** At least one index is created whose choice is justified by the query profile.
- **Preconditions:** Frequent queries are known: list selection and search, always filtered by owner, and sorting by creation date.
- **Input data:** queries `SELECT ... WHERE user_id = $1 ORDER BY created_at DESC` and owner search.
- **Expected behavior:** A composite index `(user_id, created_at)` is introduced (justification — see 05-database.md). For full name search `pg_trgm`/GIN is considered as data grows.
- **Errors:** missing index is not a runtime error, but degradation; recorded by `EXPLAIN` plan.
- **Invariants:** index does not break query correctness; used by planner for typical selections.
- **Acceptance criteria:** `EXPLAIN ANALYZE` of a typical list query uses index `(user_id, created_at)` instead of seq scan on a selection with owner filter.
- **Related Test IDs:** `TST-DB-006`.

---

# REQ-DB-004 — Transaction Atomicity

- **Description:** Operations affecting multiple steps are executed atomically; on failure no partial changes remain.
- **Preconditions:** Schema applied; DB connection available.
- **Input data:** an operation requiring a transaction (e.g., multiple changes or a step with possible failure).
- **Expected behavior:** On success — commit (`COMMIT`); on failure — rollback (`ROLLBACK`) without observable partial results.
- **Errors:** error inside transaction → rollback and `Left[E]`/`500` without leaking SQL details.
- **Invariants:** after rollback DB state is identical to state before transaction start.
- **Acceptance criteria:** a test causing failure in the middle of a transaction leaves no partial records.
- **Related Test IDs:** `TST-DB-007`.

---

# REQ-SEC-001 — Password Hashing, No Plaintext Storage

- **Description:** Passwords are hashed with Argon2id (bcrypt acceptable as fallback). Plaintext password is not stored and not logged.
- **Preconditions:** User provided a password at registration/change.
- **Input data:** plaintext password string.
- **Expected behavior:** A unique hash is created (with random salt). Comparison on login — in constant time.
- **Errors:** missing hashing algorithm — configuration error (`500`).
- **Invariants:** plaintext password does not get into DB, logs, and API responses; equal passwords give different hashes (salt).
- **Acceptance criteria:** only hash is stored in DB; `grep` of plaintext password over the table and logs finds no matches.
- **Related Test IDs:** `TST-SEC-001`, `TST-AUTH-010`.

---

# REQ-SEC-002 — JWT Signature and Lifetime, No Secret Leakage

- **Description:** JWT is signed with server secret (HMAC-SHA256 or equivalent), has a lifetime; secret is not disclosed to client.
- **Preconditions:** Secret is set in configuration (environment), not hardcoded in code.
- **Input data:** token from client; secret on server.
- **Expected behavior:** Signature is verified on each protected request; expired token is rejected. Secret is absent from any responses and static files.
- **Errors:** invalid signature/lifetime → `401` (see REQ-AUTH-003).
- **Invariants:** secret is not serialized into API responses and not committed; secret change invalidates old tokens.
- **Acceptance criteria:** token with forged signature → `401`; responses do not contain secret value.
- **Related Test IDs:** `TST-SEC-002`, `TST-SEC-003`.

---

# REQ-SEC-003 — No Internal Details Leakage in API Responses

- **Description:** API responses do not reveal internal details (SQL queries, stack traces, table names, DB driver messages).
- **Preconditions:** Any error at API boundary.
- **Input data:** failure condition (DB error, panic, invalid input).
- **Expected behavior:** Client receives only a structured error from 06-error-model.md with a safe message. Details are written to server log.
- **Errors:** leakage itself is a security defect.
- **Invariants:** `500` response body does not contain DB driver `error`, SQL, paths, and stack.
- **Acceptance criteria:** response on internal error matches error format and contains no forbidden substrings.
- **Related Test IDs:** `TST-SEC-004`.

---

# REQ-API-001 — REST Endpoints and Status Codes

- **Description:** All REST endpoints of version `/api/v1` are implemented with correct methods and status codes (see 04-api-contract.md).
- **Preconditions:** Server is running; routing is configured.
- **Input data:** HTTP requests to `auth/register`, `auth/login`, `guests`, `guests/search`, `guests/{id}`.
- **Expected behavior:** Each endpoint returns the expected status: `200/201/204` on success, `4xx` on client errors, `500` on internal ones.
- **Errors:** unknown path → `404`; unsupported method → `405`; invalid JSON → `400`.
- **Invariants:** contract is stable and versioned (`/api/v1`); successful responses match JSON schemas.
- **Acceptance criteria:** matrix "endpoint → status" is fully covered by tests.
- **Related Test IDs:** `TST-API-001`, `TST-API-004`, `TST-API-005`.

---

# REQ-API-002 — Unified Error Format

- **Description:** All errors are returned in a single structured format with code and message.
- **Preconditions:** Any failure condition at API boundary.
- **Input data:** internal error/validation/domain failure.
- **Expected behavior:** Response contains `code` (stable machine code) and `message` (human-readable, safe message). Format is unified for all endpoints.
- **Errors:** missing `code` — contract violation.
- **Invariants:** `code` matches the mapping table from 06-error-model.md; `message` does not reveal internals (REQ-SEC-003).
- **Acceptance criteria:** all tested errors return valid format with correct HTTP status.
- **Related Test IDs:** `TST-API-002`.

---

# REQ-API-003 — Input Data Validation at API Boundary

- **Description:** Input data is validated at the boundary (handler) before reaching domain/DB.
- **Preconditions:** Request with JSON body/parameters reached the endpoint.
- **Input data:** request bodies and path/query parameters.
- **Expected behavior:** Invalid input (empty full name, incorrect email, `check_out <= check_in`, empty search `q`) is rejected with `400`/`422` and a clear message. Valid data passes further into the `Either` pipeline.
- **Errors:** invalid JSON → `400`; semantically incorrect values → `422` (`VALIDATION_FAILED`).
- **Invariants:** incorrect input does not reach repository; validation errors are `Left[E]`, not panic.
- **Acceptance criteria:** a set of invalid inputs is rejected with status `400`/`422` and a list of field errors.
- **Related Test IDs:** `TST-API-003`, `TST-GUEST-002`, `TST-GUEST-003`.

---

# REQ-FE-001 — Frontend Application on TypeScript

- **Description:** Client interface implemented in TypeScript for registration, login, and guest record management (CRUD + search).
- **Preconditions:** Backend available via REST API (REQ-API-001).
- **Input data:** user actions (forms, buttons, search).
- **Expected behavior:** User can register, log in, view guest list, create/edit/delete a record, perform search. Client typing matches the API contract.
- **Errors:** API errors are shown to the user in a safe form (without internals).
- **Invariants:** client does not store plaintext passwords; requests contain `Authorization` only after login.
- **Acceptance criteria:** TypeScript project compiles (`tsc`) without errors; main scenarios are available via UI.
- **Related Test IDs:** `TST-FE-001`.

---

# REQ-FE-002 — Token Storage and Routing

- **Description:** After login JWT is saved on the client and added to requests; protected routes are available only when authenticated.
- **Preconditions:** User logged in (REQ-AUTH-002).
- **Input data:** JWT from server; navigation across routes.
- **Expected behavior:** Token is saved and substituted into `Authorization`; unauthenticated access to protected routes redirects to login page; on `401` token is reset.
- **Errors:** expired/invalid token → redirect to login page.
- **Invariants:** protected route does not render data without a valid token.
- **Acceptance criteria:** transition to protected route without token leads to login; after login routes are available.
- **Related Test IDs:** `TST-FE-002`, `TST-FE-003`.

---

# REQ-INFRA-001 — Docker Compose for Stack Deployment

- **Description:** Stack (backend, PostgreSQL, frontend) is raised with a single command via Docker Compose.
- **Preconditions:** Docker and Docker Compose installed; `.env` available based on `.env.example`.
- **Input data:** `docker-compose.yml` + environment variables.
- **Expected behavior:** `docker compose up` deploys DB, applies migrations, starts backend/frontend. Services see each other by network aliases.
- **Errors:** missing mandatory environment variable → clear startup error.
- **Invariants:** secrets are taken from environment, not from image; DB is available only within compose network.
- **Acceptance criteria:** one `docker compose up` gives a working stack; health/readiness endpoint responds.
- **Related Test IDs:** `TST-INFRA-001`.

---

# REQ-INFRA-002 — `.env.example` with Mandatory Variables

- **Description:** Repository contains `.env.example` with a list of all mandatory variables and safe default values.
- **Preconditions:** Project uses environment-based configuration.
- **Input data:** `.env.example` template.
- **Expected behavior:** At minimum listed: `DATABASE_URL`, `JWT_SECRET`, `HTTP_ADDR`, ports and DB parameters. Placeholder values are not working secrets.
- **Errors:** missing variable at runtime → configuration failure at startup.
- **Invariants:** real `.env` is not committed; `.env.example` does not contain valid secrets.
- **Acceptance criteria:** `.env.example` contains all variables read by the application; launch from template works.
- **Related Test IDs:** `TST-INFRA-002`.

---

# REQ-INFRA-003 — README with Launch Instructions

- **Description:** README describes project purpose, stack, requirements, and step-by-step launch (including tests and migrations).
- **Preconditions:** Project has reproducible launch.
- **Input data:** README text + commands.
- **Expected behavior:** A new developer deploys the project per README, applies migrations, runs tests.
- **Errors:** README discrepancy with reality — documentation defect.
- **Invariants:** commands in README are current and executable.
- **Acceptance criteria:** sections "Stack", "Launch", "Tests", "Migrations", "Environment variables" are present and executable.
- **Related Test IDs:** `TST-INFRA-003`.

---

# REQ-INFRA-004 — Test Infrastructure

- **Description:** Project is covered by tests: unit tests on `Either`/`IOEither` and domain, HTTP tests via `httptest`, integration tests with real PostgreSQL.
- **Preconditions:** Go toolchain and Docker available.
- **Input data:** `go test ./...`; integration tests are raised against a test DB.
- **Expected behavior:** Unit tests do not require DB; integration tests use isolated DB/schema and are skipped if DB unavailable (by tag/variable).
- **Errors:** any test failure — defect; integration tests without DB — skip, not false success.
- **Invariants:** tests are deterministic; integration tests leave no trash in shared DB.
- **Acceptance criteria:** `go test ./...` passes; integration tests run when DB is available and cover CRUD, isolation, constraints.
- **Related Test IDs:** `TST-INFRA-004`, `TST-INFRA-005`.

---

## Summary Traceability Table REQ → Test ID

| Requirement ID | Brief | Test ID |
|---|---|---|
| REQ-AUTH-001 | User registration | TST-AUTH-001..003, 010; TST-SEC-001 |
| REQ-AUTH-002 | Login and JWT issuance | TST-AUTH-004..006; TST-SEC-002 |
| REQ-AUTH-003 | JWT middleware | TST-AUTH-007..009; TST-SEC-002 |
| REQ-GUEST-001 | Guest creation | TST-GUEST-001..004 |
| REQ-GUEST-002 | Guest list | TST-GUEST-005, 015 |
| REQ-GUEST-003 | Guest search | TST-GUEST-006, 007; TST-DB-006 |
| REQ-GUEST-004 | Record view | TST-GUEST-008, 009, 014 |
| REQ-GUEST-005 | Editing | TST-GUEST-010, 011 |
| REQ-GUEST-006 | Deletion | TST-GUEST-012, 013 |
| REQ-GUEST-007 | Data isolation | TST-GUEST-009, 011, 013, 015 |
| REQ-ROP-001 | Either[E,T] | TST-ROP-001..003, 008..011 |
| REQ-ROP-002 | IOEither[E,T] | TST-ROP-004..007, 012..014 |
| REQ-ROP-003 | 7 ROP invariants | TST-ROP-001..007 |
| REQ-DB-001 | Migrations | TST-DB-001, 002 |
| REQ-DB-002 | FK/NOT NULL/CHECK | TST-DB-003..005 |
| REQ-DB-003 | Index | TST-DB-006 |
| REQ-DB-004 | Transactions | TST-DB-007 |
| REQ-SEC-001 | Password hashing | TST-SEC-001; TST-AUTH-010 |
| REQ-SEC-002 | JWT/secret | TST-SEC-002, 003 |
| REQ-SEC-003 | No detail leakage | TST-SEC-004 |
| REQ-API-001 | Endpoints/statuses | TST-API-001, 004, 005 |
| REQ-API-002 | Error format | TST-API-002 |
| REQ-API-003 | Boundary validation | TST-API-003; TST-GUEST-002, 003 |
| REQ-FE-001 | Frontend TS | TST-FE-001 |
| REQ-FE-002 | Token/routes FE | TST-FE-002, 003 |
| REQ-INFRA-001 | Docker Compose | TST-INFRA-001 |
| REQ-INFRA-002 | .env.example | TST-INFRA-002 |
| REQ-INFRA-003 | README | TST-INFRA-003 |
| REQ-INFRA-004 | Test infrastructure | TST-INFRA-004, 005 |
