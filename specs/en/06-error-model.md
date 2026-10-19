# 06 — Error Model (Error Model)

**System:** Guest Accounting
**Language:** English

Errors in the system are values `DomainError`, not Go exceptions.
Exceptions (`panic`) are allowed only as unexpected failures and are caught
at the HTTP layer boundary with conversion to `500`.

---

## 1. DomainError Type

```go
type DomainError struct {
    Code    string // stable machine code (contract for client)
    Message string // human-readable, safe message
    Field   string // affected field (optional, for validation)
    Err     error  // internal cause (for log, NOT for client)
}
```

Principles:
- `Code` — the only machine identifier, stable between versions.
- `Message` — safe message, without internals.
- `Field` — helps the client understand which field is invalid.
- `Err` — internal chain of causes; serialized only to the server log.

---

## 2. Error Code Reference

| Code | Category | Description | HTTP |
|---|---|---|---|
| `BAD_REQUEST` | client | invalid JSON, incorrect request syntax | `400` |
| `VALIDATION_FAILED` | client | field rule violation | `422` |
| `AUTH_TOKEN_MISSING` | auth | missing `Authorization` header | `401` |
| `AUTH_TOKEN_INVALID` | auth | invalid token signature/structure | `401` |
| `AUTH_TOKEN_EXPIRED` | auth | token lifetime expired | `401` |
| `AUTH_EMAIL_TAKEN` | auth | email already taken on registration | `409` |
| `AUTH_INVALID_CREDENTIALS` | auth | invalid email or password on login | `401` |
| `GUEST_NOT_FOUND` | domain | guest record not found or belongs to another | `404` |
| `NOT_FOUND` | transport | unknown route/resource | `404` |
| `METHOD_NOT_ALLOWED` | transport | unsupported HTTP method | `405` |
| `CONFLICT` | client | state conflicts | `409` |
| `INTERNAL` | server | unexpected error | `500` |

---

## 3. HTTP Mapper: DomainError code → HTTP status

The mapping table is used uniformly by all handlers.

| DomainError `Code` | HTTP Status | Example client message |
|---|---|---|
| `BAD_REQUEST` | `400` | Invalid request format |
| `VALIDATION_FAILED` | `422` | Invalid request data |
| `AUTH_TOKEN_MISSING` | `401` | Authentication required |
| `AUTH_TOKEN_INVALID` | `401` | Invalid token |
| `AUTH_TOKEN_EXPIRED` | `401` | Token lifetime expired |
| `AUTH_EMAIL_TAKEN` | `409` | User with such email already exists |
| `AUTH_INVALID_CREDENTIALS` | `401` | Invalid email or password |
| `GUEST_NOT_FOUND` | `404` | Record not found |
| `NOT_FOUND` | `404` | Resource not found |
| `METHOD_NOT_ALLOWED` | `405` | Method not supported |
| `CONFLICT` | `409` | Resource state conflict |
| `INTERNAL` | `500` | Internal server error |

Note on `403`: in the current model access to a guest is determined by record
ownership, and an attempt to access someone else's record returns `404` (not `403`)
so as not to reveal the existence of someone else's records (REQ-GUEST-007).
`403` remains reserved for future roles.

---

## 4. Internal Detail Hiding Rules

| Detail type | Log | Client |
|---|---|---|
| DB driver message | yes | **no** |
| SQL query text | yes | **no** |
| stack trace / panic | yes | **no** |
| table/column name | yes | **no** |
| JWT secret value | never | **no** |
| `Code` + safe `Message` | yes | **yes** |

Implementation: at the HTTP boundary `Fold`/`Map` over `Either` replaces `Err` with
a standard safe `Message` for `500`. For `4xx` the message is already safe
(given in the code reference).

Example:

```go
// inside (log):
DomainError{Code: "INTERNAL", Message: "db insert failed",
            Err: pq.Error{Code: "23505", Message: "duplicate key ..."}}

// outside (client):
{"code":"INTERNAL","message":"Internal server error"}
```

---

## 5. Error Response Format

See also `04-api-contract.md`, section 2. Unified format:

```json
{
  "code": "VALIDATION_FAILED",
  "message": "Invalid request data",
  "fields": [
    { "field": "full_name", "message": "cannot be empty" },
    { "field": "check_out", "message": "must be later than check-in date" }
  ]
}
```

`fields` is present only for validation errors (`VALIDATION_FAILED`).
For other codes `code` and `message` are sufficient.

---

## 6. Mapping from Layers

| Error source | How it becomes DomainError |
|---|---|
| domain validation (`NewGuest`) | directly `Left{VALIDATION_FAILED, Field}` |
| JWT middleware | `Left{AUTH_TOKEN_*}` |
| repository (`(T, error)`) | DB error mapping (see 05, section 7) → `Left{...}` |
| panic in handler | catch → `Left{INTERNAL}` (no details to client) |

DB errors are mapped by type:
- `unique_violation` on email → `AUTH_EMAIL_TAKEN`;
- `check_violation`/`not_null_violation` → `VALIDATION_FAILED`;
- `foreign_key_violation` → `VALIDATION_FAILED` (reference to nonexistent owner);
- others → `INTERNAL`.
