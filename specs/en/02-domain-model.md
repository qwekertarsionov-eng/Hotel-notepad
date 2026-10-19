# 02 — Domain Model (Domain Model)

**System:** Guest Accounting
**Language:** English

The domain layer describes entities, their fields, invariants, and validation rules.
The domain does not depend on HTTP, DB, and frameworks; it is expressed with pure Go types and
functions wrapped in `Either[E,T]` / `IOEither[E,T]` (see `03-architecture.md`).

---

## 1. Entity Overview

```
┌────────────┐         owner          ┌────────────┐
│   User     │1 ──────────────────── * │   Guest    │
├────────────┤                          ├────────────┤
│ Id         │                          │ Id         │
│ Email      │                          │ FullName   │
│ PasswordH  │                          │ Phone      │
│ CreatedAt  │                          │ RoomNumber │
└────────────┘                          │ CheckIn    │
                                        │ CheckOut   │
                                        │ Note       │
                                        │ UserId     │
                                        │ CreatedAt  │
                                        │ UpdatedAt  │
                                        └────────────┘
        ┌────────────┐
        │   Token    │  (JWT, computed, not a DB entity)
        ├────────────┤
        │ Subject    │  (user id)
        │ IssuedAt   │
        │ ExpiresAt  │
        └────────────┘
```

---

## 2. User Entity (User)

### 2.1 Fields

| Field | Go type | Description | Constraints |
|---|---|---|---|
| `ID` | `int64` / `uuid.UUID` | Identifier | unique, assigned by DB |
| `Email` | `string` | Email for login | unique, normalized to lowercase, valid format |
| `PasswordHash` | `[]byte` / `string` | Argon2id password hash | never contains plaintext password |
| `CreatedAt` | `time.Time` | Creation moment | set automatically |

### 2.2 User Invariants

- **INV-U-1:** `ID` is non-empty and unique within the `users` table.
- **INV-U-2:** `Email` is unique (case-insensitive after normalization).
- **INV-U-3:** `PasswordHash` is a valid Argon2id/bcrypt hash; plaintext password is absent.
- **INV-U-4:** `CreatedAt` cannot be in the future.

### 2.3 User Validation Rules

| Field | Rule | Error code |
|---|---|---|
| `Email` | non-empty, matches email format (one `@`, domain, no spaces) | `VALIDATION_FAILED` |
| `Password` | non-empty, length ≥ 8 characters | `VALIDATION_FAILED` |
| `Email` (on registration) | not yet taken | `AUTH_EMAIL_TAKEN` |

### 2.4 Domain Factory

`NewUser(email, password) -> Either[DomainError, User]`
- normalizes email (trim, lower);
- checks email format and password length;
- hashes the password (Argon2id) — side effect, therefore hashing is encapsulated in `IOEither` at the service boundary, not in the pure factory.

---

## 3. Guest Entity (Guest)

### 3.1 Fields

| Field | Go type | Description | Constraints |
|---|---|---|---|
| `ID` | `int64` / `uuid.UUID` | Identifier | unique, assigned by DB |
| `FullName` | `string` | Guest full name | non-empty, trimmed |
| `Phone` | `string` | Phone | non-empty, allowed characters (digits, `+`, spaces, hyphens, parentheses) |
| `RoomNumber` | `string` | Room number | non-empty |
| `CheckIn` | `time.Time` | Check-in date | valid date |
| `CheckOut` | `time.Time` | Check-out date | valid date **and strictly later** than `CheckIn` |
| `Note` | `string` | Note | may be empty (optional field) |
| `UserId` | `int64` / `uuid.UUID` | Owner | reference to `User.ID`, set from token |
| `CreatedAt` | `time.Time` | Creation moment | automatically |
| `UpdatedAt` | `time.Time` | Modification moment | automatically |

### 3.2 Guest Invariants

- **INV-G-1:** `ID` is non-empty and unique within the `guests` table.
- **INV-G-2:** `FullName` is non-empty after `TrimSpace`.
- **INV-G-3:** `Phone` is non-empty and contains only allowed characters.
- **INV-G-4:** `RoomNumber` is non-empty.
- **INV-G-5:** `CheckOut` is **strictly later** than `CheckIn` (`CheckOut.After(CheckIn)`).
- **INV-G-6:** `UserId` — existing user (referential integrity).
- **INV-G-7:** `UserId` cannot be changed via request body update.
- **INV-G-8:** `CheckIn`/`CheckOut` — valid calendar dates (non-zero).

### 3.3 Guest Field Validation Rules

| Field | Rule | Error code |
|---|---|---|
| `FullName` | non-empty (after trimming spaces) | `VALIDATION_FAILED` |
| `Phone` | non-empty; allowed characters: `+`, digits, space, `-`, `(`, `)` | `VALIDATION_FAILED` |
| `RoomNumber` | non-empty | `VALIDATION_FAILED` |
| `CheckIn` | valid date | `VALIDATION_FAILED` |
| `CheckOut` | valid date **and** `CheckOut > CheckIn` | `VALIDATION_FAILED` |
| `Note` | no restrictions (may be empty) | — |
| `UserId` | taken from authentication, not from body | — |

### 3.4 Domain Factories

`NewGuest(params, userId) -> Either[DomainError, Guest]`
- performs synchronous field validation (pure function, no effects);
- on any violation returns `Left[DomainError{code: VALIDATION_FAILED}]`;
- on success — `Right[Guest]` with filled `UserId`.

`ValidateGuest(g) -> Either[DomainError, Guest]`
- re-checks invariants (used on update as well).

### 3.5 Validation Examples

```
Input:  FullName="Иванов Иван", Phone="+7 999 123-45-67", Room="101",
        CheckIn=2026-05-01, CheckOut=2026-05-05
Result: Right[Guest]  (all invariants satisfied)

Input:  FullName="", ...
Result: Left[DomainError{code: VALIDATION_FAILED, field: FullName}]

Input:   CheckIn=2026-05-05, CheckOut=2026-05-01
Result: Left[DomainError{code: VALIDATION_FAILED, field: CheckOut}]
```

---

## 4. Token (JWT)

Token is not a storage entity, but a computed authentication value.

### 4.1 Structure

| Element | Description |
|---|---|
| Header | algorithm (HS256), type `JWT` |
| Payload `sub` | user identifier (`User.ID`) |
| Payload `iat` | issuance time |
| Payload `exp` | expiration time (`iat + TTL`) |
| Payload `iss` | issuer (optional) |

### 4.2 Token Invariants

- **INV-T-1:** Token is signed with server secret; signature is verified on each request.
- **INV-T-2:** `exp` is strictly in the future for a valid token.
- **INV-T-3:** Secret is not serialized into the token and is not disclosed to client.
- **INV-T-4:** `sub` is the only source of `user_id` for domain operations.

### 4.3 Token Operations

- `IssueToken(userId, secret, ttl) -> IOEither[DomainError, Token]`
- `ParseToken(tokenString, secret) -> Either[DomainError, Claims]`

---

## 5. Domain Error Type

Domain errors are values `DomainError`, not exceptions. Full model is in
`06-error-model.md`. Field summary:

```go
type DomainError struct {
    Code    string // stable machine code, e.g. VALIDATION_FAILED
    Message string // human-readable, safe message
    Field   string // affected field (optional)
    Err     error  // internal cause (not sent to client)
}
```

Domain codes used at the model level:

| Code | Application |
|---|---|
| `VALIDATION_FAILED` | User/Guest field invariant violation |
| `AUTH_EMAIL_TAKEN` | email already taken on registration |
| `AUTH_INVALID_CREDENTIALS` | invalid credentials on login |
| `GUEST_NOT_FOUND` | record not found or belongs to another |
| `NOT_FOUND` | resource not found |
| `INTERNAL` | unexpected error |

---

## 6. Domain vs DB Correspondence

| Domain | Table | See `05-database.md` |
|---|---|---|
| `User` | `users` | `NOT NULL` constraints, `UNIQUE(email)` |
| `Guest` | `guests` | `FK(user_id)`, `NOT NULL`, `CHECK(check_out > check_in)` |
| `Token` | — | not stored (stateless JWT) |

---

## 7. Domain Invariant Summary

| ID | Entity | Invariant |
|---|---|---|
| INV-U-1 | User | ID is unique |
| INV-U-2 | User | Email is unique (normalized) |
| INV-U-3 | User | only password hash is stored |
| INV-U-4 | User | CreatedAt is not in the future |
| INV-G-1 | Guest | ID is unique |
| INV-G-2 | Guest | FullName is non-empty |
| INV-G-3 | Guest | Phone is non-empty and of allowed characters |
| INV-G-4 | Guest | RoomNumber is non-empty |
| INV-G-5 | Guest | CheckOut > CheckIn |
| INV-G-6 | Guest | UserId exists |
| INV-G-7 | Guest | UserId is not changed via request body |
| INV-G-8 | Guest | dates are valid |
| INV-T-1..4 | Token | signature, lifetime, secret, user_id source |
