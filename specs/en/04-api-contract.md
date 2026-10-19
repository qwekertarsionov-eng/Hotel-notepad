# 04 — API Contract (API Contract)

**System:** Guest Accounting
**Base prefix:** `/api/v1`
**Data format:** JSON (UTF-8 encoding, `Content-Type: application/json`)
**Authentication:** JWT in header `Authorization: Bearer <token>`

Error codes and their HTTP statuses are in `06-error-model.md`.
Mapping `DomainError` → HTTP is performed uniformly.

---

## 1. Endpoint Summary Table

| Method | Path | Description | Auth | Success |
|---|---|---|---|---|
| POST | `/api/v1/auth/register` | User registration | no | `201` |
| POST | `/api/v1/auth/login` | Login, JWT issuance | no | `200` |
| GET | `/api/v1/guests` | Guest list (pagination) | yes | `200` |
| POST | `/api/v1/guests` | Guest creation | yes | `201` |
| GET | `/api/v1/guests/search` | Guest search | yes | `200` |
| GET | `/api/v1/guests/{id}` | Record view | yes | `200` |
| PUT | `/api/v1/guests/{id}` | Record editing | yes | `200` |
| DELETE | `/api/v1/guests/{id}` | Record deletion | yes | `204` |

Status codes used in the contract: `200`, `201`, `204`, `400`, `401`,
`403`, `404`, `409`, `422`, `500`.

---

## 2. Unified Error Format

Any error returns JSON:

```json
{
  "code": "VALIDATION_FAILED",
  "message": "Invalid request data",
  "fields": [
    { "field": "check_out", "message": "check-out date must be later than check-in date" }
  ]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `code` | string | yes | stable machine code (see 06-error-model.md) |
| `message` | string | yes | human-readable, safe message |
| `fields` | array | no | field details (for validation errors) |

Rule: `500` response body does not contain SQL, stack traces, paths, and DB driver messages (REQ-SEC-003).

---

## 3. Authentication

### 3.1 `POST /api/v1/auth/register`

Request:
```json
{
  "email": "user@example.com",
  "password": "secret-pass-1"
}
```

Response `201 Created`:
```json
{
  "id": 1,
  "email": "user@example.com"
}
```

Errors:

| Status | code | Condition |
|---|---|---|
| `409` | `AUTH_EMAIL_TAKEN` | email already taken |
| `422` | `VALIDATION_FAILED` | invalid email or short password |
| `500` | `INTERNAL` | DB failure |

### 3.2 `POST /api/v1/auth/login`

Request:
```json
{
  "email": "user@example.com",
  "password": "secret-pass-1"
}
```

Response `200 OK`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Errors:

| Status | code | Condition |
|---|---|---|
| `401` | `AUTH_INVALID_CREDENTIALS` | invalid email or password |
| `422` | `VALIDATION_FAILED` | empty fields |
| `500` | `INTERNAL` | failure |

Rule: response for nonexistent email and for wrong password is identical, so as not to reveal account existence.

---

## 4. Guest Resource

All endpoints below require a valid JWT. Otherwise → `401`.

### 4.1 Guest Object Schema (response)

```json
{
  "id": 10,
  "full_name": "Иванов Иван Иванович",
  "phone": "+7 999 123-45-67",
  "room_number": "101",
  "check_in": "2026-05-01",
  "check_out": "2026-05-05",
  "note": "early check-in",
  "user_id": 1,
  "created_at": "2026-04-25T10:00:00Z",
  "updated_at": "2026-04-25T10:00:00Z"
}
```

### 4.2 `POST /api/v1/guests` — creation

Request (body):
```json
{
  "full_name": "Иванов Иван Иванович",
  "phone": "+7 999 123-45-67",
  "room_number": "101",
  "check_in": "2026-05-01",
  "check_out": "2026-05-05",
  "note": "early check-in"
}
```

`user_id` **is not passed** in the body — taken from the token.

Response `201 Created` — full Guest object.

Errors:

| Status | code | Condition |
|---|---|---|
| `401` | `AUTH_TOKEN_*` | missing/broken/expired token |
| `400` | `BAD_REQUEST` | invalid JSON |
| `422` | `VALIDATION_FAILED` | empty full name, `check_out <= check_in`, etc. |
| `500` | `INTERNAL` | DB failure |

### 4.3 `GET /api/v1/guests` — list

Query parameters:

| Parameter | Type | Default | Description |
|---|---|---|---|
| `page` | int ≥ 1 | `1` | page number |
| `limit` | int (1..100) | `20` | page size |

Response `200 OK`:
```json
{
  "items": [ /* array of Guest */ ],
  "page": 1,
  "limit": 20,
  "total": 42
}
```

Errors: `400` on incorrect pagination; `401` without token.

### 4.4 `GET /api/v1/guests/search` — search

Query parameters:

| Parameter | Type | Required | Description |
|---|---|---|---|
| `q` | string (non-empty) | yes | search string |
| `field` | `full_name`/`phone`/`room_number`/`all` | no | search field (default `all`) |

Response `200 OK` — as in the list (`items` + metadata). Empty result —
`200` with empty `items`, not an error.

Errors: `400` on empty `q`; `401` without token.

### 4.5 `GET /api/v1/guests/{id}` — view

Response `200 OK` — Guest object.

Errors:

| Status | code | Condition |
|---|---|---|
| `401` | `AUTH_TOKEN_*` | missing/broken token |
| `404` | `GUEST_NOT_FOUND` | record not found **or** belongs to another |

Someone else's record and a nonexistent record are indistinguishable to the client.

### 4.6 `PUT /api/v1/guests/{id}` — editing

Request (body) — same fields as on creation (`full_name`, `phone`,
`room_number`, `check_in`, `check_out`, `note`). `id` and `user_id` are ignored
if passed.

Response `200 OK` — updated Guest object.

Errors: `401`; `404` (`GUEST_NOT_FOUND`); `422` (`VALIDATION_FAILED`); `500`.

### 4.7 `DELETE /api/v1/guests/{id}` — deletion

Response `204 No Content` (no body).

Errors: `401`; `404` (`GUEST_NOT_FOUND`); `500`.

---

## 5. Unknown Routes and Methods

| Condition | Status | code |
|---|---|---|
| Unknown path | `404` | `NOT_FOUND` |
| Known path, unsupported method | `405` | `METHOD_NOT_ALLOWED` |
| Invalid JSON in body | `400` | `BAD_REQUEST` |

---

## 6. Status Matrix

| Endpoint | 200 | 201 | 204 | 400 | 401 | 404 | 409 | 422 | 500 |
|---|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|:-:|
| `POST /auth/register` | | ✓ | | | | | ✓ | ✓ | ✓ |
| `POST /auth/login` | ✓ | | | | ✓ | | | ✓ | ✓ |
| `GET /guests` | ✓ | | | ✓ | ✓ | | | | ✓ |
| `POST /guests` | | ✓ | | ✓ | ✓ | | | ✓ | ✓ |
| `GET /guests/search` | ✓ | | | ✓ | ✓ | | | | ✓ |
| `GET /guests/{id}` | ✓ | | | | ✓ | ✓ | | | ✓ |
| `PUT /guests/{id}` | ✓ | | | ✓ | ✓ | ✓ | | ✓ | ✓ |
| `DELETE /guests/{id}` | | | ✓ | | ✓ | ✓ | | | ✓ |

Note: `403` is reserved for future role separation; in the current model
access is determined by record ownership (isolation via `404`, see REQ-GUEST-007).

---

## 7. Date Format

- Dates (`check_in`, `check_out`) are passed in `YYYY-MM-DD` format (calendar date).
- Time points (`created_at`, `updated_at`) — ISO 8601 / RFC 3339 (`YYYY-MM-DDTHH:MM:SSZ`).
