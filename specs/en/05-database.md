# 05 — Database (Database)

**System:** Guest Accounting
**DBMS:** PostgreSQL
**Default schema:** `public`

---

## 1. Table Overview

| Table | Purpose |
|---|---|
| `users` | user accounts (email + password hash) |
| `guests` | guest records belonging to a user |
| `schema_migrations` | service table of migration versions |

Relation: `guests.user_id → users.id` (many guests — one user).

---

## 2. `users` Table

```sql
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
```

| Column | Type | Constraints | Rationale |
|---|---|---|---|
| `id` | `BIGSERIAL` | `PRIMARY KEY` | unique record identifier |
| `email` | `VARCHAR(255)` | `NOT NULL`, `UNIQUE` | login, normalized case-insensitively before write |
| `password_hash` | `TEXT` | `NOT NULL` | Argon2id hash, plaintext password not stored |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | audit |

```sql
CREATE UNIQUE INDEX ux_users_email ON users (email);
```

`UNIQUE(email)` guarantees INV-U-2. Normalization (lower/trim) is performed in
the application before insertion, so uniqueness is deterministic.

---

## 3. `guests` Table

```sql
CREATE TABLE guests (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    full_name   VARCHAR(255) NOT NULL,
    phone       VARCHAR(50)  NOT NULL,
    room_number VARCHAR(50)  NOT NULL,
    check_in    DATE         NOT NULL,
    check_out   DATE         NOT NULL,
    note        TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT fk_guests_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_guests_dates
        CHECK (check_out > check_in)
);
```

| Column | Type | Constraints | Rationale |
|---|---|---|---|
| `id` | `BIGSERIAL` | `PRIMARY KEY` | unique record identifier |
| `user_id` | `BIGINT` | `NOT NULL`, `FK → users.id` | owner; referential integrity (INV-G-6) |
| `full_name` | `VARCHAR(255)` | `NOT NULL` | INV-G-2 |
| `phone` | `VARCHAR(50)` | `NOT NULL` | INV-G-3 |
| `room_number` | `VARCHAR(50)` | `NOT NULL` | INV-G-4 |
| `check_in` | `DATE` | `NOT NULL` | INV-G-8 |
| `check_out` | `DATE` | `NOT NULL`, `CHECK (check_out > check_in)` | INV-G-5 |
| `note` | `TEXT` | `NOT NULL DEFAULT ''` | optional field |
| `created_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | audit, sorting |
| `updated_at` | `TIMESTAMPTZ` | `NOT NULL DEFAULT now()` | audit |

### 3.1 Foreign Key

`fk_guests_user`: `guests.user_id → users.id`.
`ON DELETE CASCADE` — when a user is deleted, their guests are deleted too
(the alternative `ON DELETE RESTRICT` — deletion prohibition; cascade is chosen
as more natural behavior for accounting). Nonexistent `user_id` cannot be written.

### 3.2 CHECK Constraint on Dates

`chk_guests_dates`: `CHECK (check_out > check_in)`.
Insurance above application validation (REQ-DB-002): even when bypassing the API
it is impossible to save check-out earlier than check-in.

---

## 4. Index Justification (minimum one)

### 4.1 Choice: composite index `(user_id, created_at)`

```sql
CREATE INDEX ix_guests_user_created_at ON guests (user_id, created_at DESC);
```

**Query profile (justification of choice):**
- Guest list: `SELECT ... WHERE user_id = $1 ORDER BY created_at DESC LIMIT ...` (REQ-GUEST-002).
- Guest search: `SELECT ... WHERE user_id = $1 AND <filter>` (REQ-GUEST-003).
- All isolation operations filter by `user_id` (REQ-GUEST-007).

**Why composite, not only `user_id`:**
An index with leading `user_id` serves the owner filter, while the leading order
`created_at DESC` allows returning pages in the required order without additional
sorting (sorting by index instead of `Sort` in the plan). This closes both
filtering and sorting with a single index — a typical optimization technique
"sorting via index".

### 4.2 Alternative: `pg_trgm` / GIN on full name

For full name substring search (`ILIKE '%substring%'`) a leading B-tree index does
not help (predicate with a free prefix). Optionally:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX ix_guests_full_name_trgm ON guests USING gin (full_name gin_trgm_ops);
```

**Decision:** at the current volume (study/early product) the composite
`(user_id, created_at)` is chosen as the main and justified index. `pg_trgm` —
a scalable fallback option as records grow and full name search becomes frequent;
it is fixed here as a conscious improvement, not as an MVP requirement.
Note: `gin_trgm_ops` index is updated on each insert — a compromise between
write and search.

### 4.3 Planner Choice Justification (check)

Index correctness is confirmed by `EXPLAIN ANALYZE` of a typical query
(REQ-DB-003, test `TST-DB-006`): the plan uses `ix_guests_user_created_at`
(Index Scan / Index Scan Backward) instead of Seq Scan when there is a selection
by `user_id`.

---

## 5. Migrations

### 5.1 File Structure

```
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_guests.up.sql
├── 000002_create_guests.down.sql
├── 000003_add_guests_indexes.up.sql
└── 000003_add_guests_indexes.down.sql
```

### 5.2 Migration Rules

| Rule | Description |
|---|---|
| Versioning | files named `NNNNNN_description.up/down.sql` |
| Order | applied in ascending version order |
| Rollback | each migration has a `.down.sql` |
| Idempotency | re-application of an already applied migration does not break the schema (tracked in `schema_migrations`) |
| Atomicity | each migration is executed in a transaction |

### 5.3 Service Table

Managed by the migration tool (e.g. `golang-migrate`): stores the current version,
prevents re-application.

### 5.4 Migration example `000002_create_guests.up.sql`

```sql
CREATE TABLE guests (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL,
    full_name   VARCHAR(255) NOT NULL,
    phone       VARCHAR(50)  NOT NULL,
    room_number VARCHAR(50)  NOT NULL,
    check_in    DATE         NOT NULL,
    check_out   DATE         NOT NULL,
    note        TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT fk_guests_user
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_guests_dates
        CHECK (check_out > check_in)
);
```

---

## 6. Transactions (REQ-DB-004)

- Multi-step operations are executed in a single transaction.
- Success → `COMMIT`; any failure → `ROLLBACK` without partial changes.
- `context.Context` is passed to the connection for timeout/cancellation.
- The adapter layer encapsulates `Begin/Commit/Rollback`; the service does not manage transactions directly.

Example boundaries (pseudo):

```go
func (r *postgresGuestRepo) Insert(ctx context.Context, g Guest) (Guest, error) {
    // tx := Begin(ctx); defer tx.Rollback()
    // INSERT ... ; return converted error as (Guest, error)
}
```

---

## 7. DB Error Mapping

| PostgreSQL error | Mapping | HTTP (see 06) |
|---|---|---|
| `unique_violation` (email) | `AUTH_EMAIL_TAKEN` | `409` |
| `foreign_key_violation` | `VALIDATION_FAILED`/`INTERNAL` | `422`/`500` |
| `check_violation` (dates) | `VALIDATION_FAILED` | `422` |
| `not_null_violation` | `VALIDATION_FAILED` | `422` |
| others (`connection`, timeout) | `INTERNAL` | `500` |

Important: technical details (`error.Error()` of the driver) are not sent to the client —
written to log, only a safe message goes outside (REQ-SEC-003).

---

## 8. Summary of Constraints

| Type | Object | Purpose |
|---|---|---|
| PK | `users.id`, `guests.id` | record uniqueness |
| UNIQUE | `users.email` | INV-U-2 |
| NOT NULL | all mandatory columns | INV-G-2..4, INV-G-8, INV-U-3 |
| FK | `guests.user_id → users.id` | INV-G-6, isolation |
| CHECK | `guests.check_out > check_in` | INV-G-5 |
| INDEX | `guests(user_id, created_at DESC)` | list/search performance |
