# 09 — TOC Analysis (Theory of Constraints)

**System:** Guest Accounting
**Language:** English
**Status:** Approved (CLR defect fixes, TASK-008)
**TOC tools:** CRT, Evaporating Cloud, FRT (with Negative Branch), PRT,
Transition Tree, Strategy & Tactics Tree

All diagrams — ASCII, monospaced font, arrows `+-->`.
Each causal link is formulated as `A --> B (B because A)`.

---

## 0. Notation, Link Format, and Life Cycle Stages

### 0.1 Terms

| Term | Expansion |
|---|---|
| UDE | Undesirable Effect — undesirable effect of the current state |
| DCE | Desirable Core Entity — in FRT this is DE0 (desired effect); CRT root is denoted D1 and is the core problem, not DCE |
| Injector | injection — an introduced change that removes the cause, not the symptom |
| Negative Branch | undesirable side branch of FRT, verified and closed by remediation |
| Lifter | positive stimulus — an unexpected beneficial effect of an injection |
| SDD | Spec-Driven Development — specifications `01..08` as the source of truth until code |
| CLR | Code Layout Review — code layout review: `go.mod`, directories per 03 §10, layers, port/adapter |
| Translation | future stage: create and synchronize `specs/en/` based on the frozen `ru` base, preserving REQ/Test/GATE IDs (directory `specs/en/` not yet created) |
| TDD L1 | TDD level 1: unit cycles "red → green" without DB (ROP types, domain) |
| Stubs | stub adapters (in-memory repository) for testing the service layer without DB |
| GATE-1 | traceability gateway: each REQ ≥ 1 Test ID and no "hanging" Test ID (08 §2) |
| GATE-2 | build gateway: `go build ./...` and `go vet` without errors (08 §2) |
| TDD L2 | TDD level 2: HTTP tests `httptest` + integration tests with PostgreSQL |
| Refactor | refactoring only on green tests, without changing observable behavior |
| Security-Core | security kernel and schema stage: Argon2id, JWT HS256, middleware, SQL migrations, unified error format, use-cases/handlers (injections I3, I4, I5, I8) |
| Security/Infra | REQ-INFRA-*, REQ-FE-* stage: Docker Compose, .env.example, README, frontend TS (without auth — it is in Security-Core) |
| GATE-3 | gateway: `go test ./...` fully green (08 §2) |

### 0.2 Link Format

Any causal link is recorded in two equivalent ways:

```
A +--> B        and        A --> B (B because A)
```

Under each scheme, the "Links" list provides the full form with justification.
Link `A x C` denotes mutually exclusive (conflicting) nodes.

### 0.3 Life Cycle Stages (used in PRT, Transition Tree, S&T)

```
SDD --> CLR --> TDD L1 --> Stubs --> GATE-1 --> Translation --> Security-Core
  --> TDD L2 --> Refactor --> Security/Infra --> GATE-3
```

Initial state (per `reports/repository-analysis.md`): directory `go/` is empty —
no `.go` files, `go.mod`, tests, Compose, README; `specs/ru/01..08` approved
(SDD passed; DB schema 05 and API contract 04 exist). CRT describes the
subject problem — absence of the guest accounting system and its risks, — not
the absence of documentation. Further — CLR and subsequent stages.

---

## 1. CRT — Current Reality Tree

Current state: greenfield project without a guest accounting system. D1 — the
single root problem; all U-nodes are its consequences; U1 — the final UDE.
Causality direction is uniform throughout the tree: `D1 --> U-nodes --> U1`.

### 1.1 Scheme

```
                        ┌──────────────────────────────────────────────────┐
                        │ D1 (root problem): missing                       │
                        │ versioned, testable foundation of the            │
                        │ accounting system — the link "requirement → code│
                        │ → test" is not implemented                       │
                        └───┬───────┬───────┬───────┬───────┬───────┬──────┘
                            │       │       │       │       │       │
               ┌────────────┤       │       │       │       │       ├────────────┐
               ▼            ▼       ▼       ▼       ▼       ▼       ▼            ▼
         ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
         │ U2: no   │ │ U3: no   │ │ U4: no   │ │ U5: no   │ │U10: no   │ │U11: no   │
         │ code and │ │ DB       │ │ API-     │ │ ROP-     │ │JWT-mid-  │ │domain    │
         │ structure│ │ schema   │ │ contract │ │ abstrac- │ │dleware   │ │factories │
         │ of       │ │ and mi-  │ │ and lay- │ │ tions    │ │and filter│ │validation│
         │ project  │ │ grations │ │ ers back-│ │          │ │user_id   │ │          │
         └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
              │            │            │            │            │            │
              │            │            │            │            │            │
              │            │            │            │            │            │
              │            │            │            ▼            ▼            │
              │            │            │      ┌──────────┐ ┌──────────┐       │
              │            │            │      │ U6: no   │ │ U7: no   │       │
              │            │            │      │ JWT-authen-│ │validation│       │
              │            │            │      │ tication   │ │ at edge  │       │
              │            │            │      │ and iso-   │ │ and sin- │       │
              │            │            │      │ lation user│ │ gle      │       │
              │            │            │      │           │ │ format   │       │
              │            │            │      └────┬─────┘ └────┬─────┘       │
              │            │            │           │            │            │
              │            │            │           │            │            │
              ▼            ▼            ▼           ▼            ▼            ▼
         ┌────────────────────────────────────────────────────────────────────────┐
         │ U1 (UDE): no working guest accounting system — nowhere to enter,       │
         │ search and store guests; MVP cannot be demonstrated, accepted          │
         │ and entrusted to users                                                 │
         └────────────────────────────────────────────────────────────────────────┘
         ┌────────────────────────┐   ┌──────────────────────────────────────────┐
         │ U8: no test infra-     │   │ U9: no reproducible launch                 │
         │ structure →            │   │ (Compose, README, .env) — "works          │
         │ regressions unnoticed  │   │ for the author"                          │
         └────────────────────────┘   └──────────────────────────────────────────┘
         (U8 and U9 — direct consequences of D1, enter U1 via branches §1.2)
```

### 1.2 Links (A --> B (B because A))

```
D1 --> U2  (U2 because D1: without a verifiable foundation there is neither a
            build unit nor the criterion "layout is correct" — code does not start)
D1 --> U3  (U3 because D1: schema outside the link "REQ → migration → test"
            is not verifiable and is not created)
D1 --> U4  (U4 because D1: contract and layers are designed only as a
            verifiable part of the foundation "requirement → code → test")
D1 --> U5  (U5 because D1: ROP types are an element of the verifiable foundation;
            without it errors remain exceptions outside composition)
D1 --> U10 (U10 because D1: auth infrastructure is laid into the system foundation,
            not added "later" to unverifiable code)
D1 --> U11 (U11 because D1: domain invariants are fixed in verifiable factories
            of the foundation, not validated by chance)
D1 --> U8  (U8 because D1: test infrastructure is a verifiable part of the foundation;
            without it regressions are indistinguishable from correct behavior)
D1 --> U9  (U9 because D1: reproducible launch is a verifiable part of the foundation;
            without it the work result is not reproducible)
U5 --> U6  (U6 because U5 partially: without Either/IOEither service failures do not
            compose — isolation failure is indistinguishable from a normal failure)
U10 --> U6 (U6 because U10: isolation REQ-GUEST-007 is ensured by JWT middleware
            and the user_id filter, not by ROP types)
U5 --> U7  (U7 because U5 partially: without a single DomainError validation failures
            do not compose predictably across the whole chain — but validation itself
            lives in factories U11)
U11 --> U7 (U7 because U11: cutting off invalid input (empty full name,
            check_out <= check_in) before DB — work of domain factories NewUser/NewGuest)
U2 --> U1  (U1 because U2: without go.mod, directories and code there is nothing to write and run)
U3 --> U1  (U1 because U3: without schema users/guests and migrations reliable storage is impossible)
U4 --> U1  (U1 because U4: without contract and layers handler/service/repository there is no addressable API)
U6 --> U1  (U1 because U6: without JWT and user_id filter the system cannot be used
            by several users)
U7 --> U1  (U1 because U7: without boundary validation incorrect data gets into DB,
            the system cannot be trusted)
U8 --> U1  (U1 because U8: acceptance without tests is subjective and allows hidden
            regressions — MVP cannot be entrusted to users)
U9 --> U1  (U1 because U9: without reproducible launch acceptance and project handover are impossible)
```

### 1.3 Explanation

CRT describes not "bad code" (there is no code) and not the absence of documentation
(specifications `01..08` are approved), but **the absence of a verifiable link
"requirement → code → test"**: traceability REQ → Test exists on paper,
GATE-1 has not been run, not a single `.go` file exists. D1 — the single root
problem formulated in one phrase; all eight original UDEs (U2..U9) and two
added ones (U10, U11) are its consequences, and U1 — the final UDE gathering
the branches. Treating UDEs one by one = symptomatic treatment.

Key conclusion of CRT: **development speed does not depend on the volume of CRUD,
but on the absence of a foundation** — ROP types, schema with constraints, JWT isolation,
and test feedback. Therefore the first investments should go into the foundation
(see FRT, injections I1..I8), not into endpoints.

### 1.4 Connection to Requirements

| Node | REQ |
|---|---|
| U3 | REQ-DB-001, REQ-DB-002, REQ-DB-003, REQ-DB-004 |
| U4 | REQ-API-001, REQ-API-002, REQ-API-003 |
| U5 | REQ-ROP-001, REQ-ROP-002, REQ-ROP-003 |
| U6 | REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003, REQ-GUEST-007, REQ-SEC-001, REQ-SEC-002 |
| U7 | REQ-API-003, REQ-GUEST-001 (date invariant) |
| U8 | REQ-INFRA-004 |
| U9 | REQ-INFRA-001, REQ-INFRA-002, REQ-INFRA-003 |
| U10 | REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003, REQ-GUEST-007 |
| U11 | REQ-GUEST-001, REQ-GUEST-005, REQ-API-003 |
| U1 | REQ-GUEST-001..006 (CRUD not yet implemented), REQ-FE-001, REQ-FE-002 |

---

## 2. Evaporating Cloud — Conflict Cloud (Conflict Diagram)

Conflict: **development speed** against **quality/safety/strictness of SDD→TDD**.
The traditional compromise ("fast now, debt later") is the undesirable state UDE-C1
(not to be confused with U1 CRT).

### 2.1 Scheme

```
                         ┌────────────────────────────────────────────┐
                         │ N1 (inviolable need):                      │
                         │ guest accounting system must exist         │
                         │ and be reliable — data (guests, passwords, │
                         │ someone else's records) cannot be trusted  │
                         │ "approximately"                            │
                         └────────────────────┬───────────────────────┘
                                              │ N1 --> C1 (C1 because N1)
                         ┌────────────────────▼───────────────────────┐
                         │ C1 (object / common goal):                 │
                         │ build an MVP of guest accounting system,   │
                         │ which can be accepted and trusted          │
                         └───────┬─────────────────────────┬──────────┘
               C1 --> C2         │                         │         C1 --> C3
                 ▼               │                         │               ▼
         ┌──────────────┐        │                         │      ┌──────────────┐
         │ C2 (need A): │        │                         │      │ C3 (need B): │
         │ maximize   │        │                         │      │ ensure       │
         │ speed of   │        │                         │      │ quality and  │
         │ getting    │        │                         │      │ safety:      │
         │ working    │        │                         │      │ SDD→TDD      │
         │ CRUD       │        │                         │      │ strictness   │
         └──────┬───────┘        │                         │      └──────┬───────┘
                │ C2 --> C4      │                         │      C3 --> C5
                ▼                │                         │             ▼
         ┌──────────────┐        │  C4 x C5 — CONFLICT     │      ┌──────────────┐
         │ C4 (action   │        │  "write quickly"        │      │ C5 (action │
         │ A): "write  │───────▶│  x "write strictly      │◀─────│ B): "write │
         │ quickly"   │        │  according to SDD/TDD"    │      │ strictly"   │
         └──────┬───────┘        └────────────┬────────────┘      └──────┬───────┘
                │ C4 --> C6                   │ C5 --> C6                │
                └─────────────────────────────┼──────────────────────────┘
                                              ▼
                         ┌────────────────────────────────────────────┐
                         │ C6 (conflict realized): in one stream of   │
                         │ work simultaneous "quickly" and "strictly" │
                         │ is impossible while strictness is perceived│
                         │ as external delay (assumption A1)          │
                         └────────────────────┬───────────────────────┘
                                              │ C6 + A1 --> UDE-C1
                                              ▼
                         ┌────────────────────────────────────────────┐
                         │ UDE-C1 (undesirable state):                │
                         │ compromise "quickly now, debts later":     │
                         │ regressions, leaks of someone else's data, │
                         │ unstructured errors, distrust of changes   │
                         │ (= UDE CRT)                                │
                         └────────────────────────────────────────────┘

        C5 --> UDE-C2 (UDE-C2 because C5: strict path in a short sprint
                       without removing assumption A1 leads to deadline miss —
                       a separate undesirable state, not the UDE-C1 compromise)
```

### 2.2 Links (A --> B (B because A))

```
N1 --> C1 (C1 because N1: a reliable system must be built; the goal is MVP)
C1 --> C2 (C2 because C1: users need a fast first working CRUD)
C1 --> C3 (C3 because C1: acceptance is impossible without integrity, isolation, predictability)
C2 --> C4 (C4 because C2: maximizing speed pulls toward minimizing stops
           "on invariants/types/tests" — write "first, then finish later")
C3 --> C5 (C5 because C3: safety and correctness require strictness SDD→TDD:
           fixed invariants, Either/IOEither, tests before/during code)
C4 x C5  (conflict: "write quickly" and "write strictly" are incompatible in one
          stream while strictness is perceived as external delay)
C4 --> C6 (C6 because C4: "write quickly" is incompatible with "write strictly"
           in one work stream)
C5 --> C6 (C6 because C5: "write strictly" is incompatible with "write quickly"
           in one work stream)
C6 --> UDE-C1 (UDE-C1 because C6 + assumption A1 "strictness = delay":
               the realized conflict forces the compromise "quickly now, debts later")
C5 --> UDE-C2 (UDE-C2 because C5: strict path in a short sprint without removing
               assumption A1 leads to deadline miss — a separate undesirable state)
```

C6 is defined explicitly: this is the node of conflict realization, which includes
**both** sides (C4 and C5). The cause of UDE-C1 is the joint action of C4 × C5 through
C6 under assumption A1, not each side separately.

### 2.3 Conflict Evaporation (conflict removal, not compromise)

Evaporation question: *"Is strictness SDD→TDD necessarily a delay relative to speed?"*
Answer: **no** — assumption A1 is destroyed (see S&T): strictness embedded into the
foundation from the first commit **reduces** total time.

Injections removing the conflict:

| Injection | How it removes C4 x C5 | REQ |
|---|---|---|
| I1: ROP types written once | speed up all subsequent use-cases: no manual branched error handling → "quickly" and "strictly" coincide | REQ-ROP-001..003 |
| I3: JWT middleware + `user_id` isolation written once | spread to all endpoints without repeat work → security "free" on each new route | REQ-AUTH-001..003, REQ-GUEST-007 |
| I4/I6: migrations and test skeletons from the first commit | feedback is immediate (test fails right away), not "at the end of the sprint" → speed due to absence of rework | REQ-DB-001..004, REQ-INFRA-004 |
| I8: use-cases and handlers written on top of the ready foundation | each next CRUD scenario is composition of ready layers, not new implementation from scratch | REQ-GUEST-001..006, REQ-API-001 |
| SDD as source of truth | specification approved before code → behavior disputes resolved by reference to REQ, not by rewriting | REQ-API-001..003 |

Alternative causes of UDE-C1 (checked and rejected):

| Alternative cause | Why rejected |
|---|---|
| Undefined scope | MVP scope is fixed by 29 REQ in 01; changes — only via REQ revision, not as a hidden cause of compromise |
| Inexperience of team | ROP/SDD reduce experience requirements: behavior is set by REQ and tests; checked by GATE-1/GATE-3, not by seniority |
| Changing requirements | Specification is frozen by traceability to GATE-1; requirement change is formalized as REQ revision and does not generate the compromise "debts later" |

Destroyed assumptions:

- **Against C4:** "fast CRUD without ROP/types/tests is really faster" — refuted by the risk of someone else's data leak (REQ-GUEST-007) and "gloomy" error handling, which are more expensive at the rework stage than at the foundation stage.
- **Against C5:** "strictness = slowdown" — refuted by the fact that GATE-1 (traceability) and GATE-3 (green tests) speed up acceptance, excluding manual regression hunting.

### 2.4 Clarifying Questions

1. What is more expensive on the project horizon: write invariants and tests immediately or eliminate incidents of someone else's data leak and regressions after?
2. Is TDD necessarily postponed for a deadline if the "test → code" cycle on ROP types takes minutes, not days?
3. Can security (isolation, error format) be made a one-time infrastructure rather than repeated work in each endpoint?
4. Does SDD strictness slow down after GATE-1, or before it — is this exactly the investment that pays off on the second and subsequent use-cases?
5. What blocks acceptance right now: absence of code or absence of verifiable criteria? (By CRT — the second: GATE-1/GATE-3 have not been run.)

### 2.5 Connection to Requirements

REQ-ROP-001..003 (speed injection through types), REQ-AUTH-001..003 and
REQ-GUEST-007 (security as one-time infrastructure), REQ-DB-001..004
(schema safety nets), REQ-API-001..003 (contract removes behavior disputes),
REQ-GUEST-001..006 (use-cases on top of foundation), REQ-INFRA-004 (test feedback
as a source of speed).

---

## 3. FRT — Future Reality Tree

Target state after implementing injections I1..I8. Negative Branch verification
and Lifter check.

### 3.1 Scheme

```
 ┌───────────────────────────────────────────────────────────────────────────────┐
 │ Injections (Injectors) — each bound to REQ:                                    │
 │  I1: Either/IOEither + 7 ROP invariants              (REQ-ROP-001..003)        │
 │  I2: User/Guest domain, factories, date validation   (REQ-GUEST-001,005,      │
 │                                                            REQ-API-003)       │
 │  I3: JWT middleware + Argon2id + secret from env     (REQ-AUTH-001..003,      │
 │                                                            REQ-SEC-001,002)   │
 │  I4: Migrations, FK/NOT NULL/CHECK, index            (REQ-DB-001..004)        │
 │  I5: Unified error format, boundary validation       (REQ-API-001..003,      │
 │                                                            REQ-SEC-003)       │
 │  I6: Tests: unit + httptest + integration with DB    (REQ-INFRA-004)          │
 │  I7: Docker Compose + .env.example + README + FE TS  (REQ-INFRA-001..003,    │
 │                                                            REQ-FE-001,002)    │
 │  I8: GuestService/AuthService use-cases + HTTP       (REQ-GUEST-001..006,    │
 │        handlers, handler → service → repository link  REQ-API-001)           │
 └───────────────────────────────────────────────────────────────────────────────┘

  I1 +--> DE1: errors are values; Either/IOEither composition is predictable,
   |            errors are not lost and do not surface as Go exceptions
   |
   I2 +--> DE2: invalid input (empty full name, check_out <= check_in)
   |            is filtered before DB — at domain and API boundary levels
   |
   I3 +--> DE3: routes are protected; user_id taken only from token;
   |            passwords stored only as hash; someone else's record
   |            is indistinguishable from nonexistent (404)
   |
   I4 +--> DE4: data integrity (FK, CHECK) and performance
   |            (index user_id, created_at) ensured at DB level
   |
   I5 +--> DE5: client receives a stable safe error contract
   |            (code + message, no SQL/stack/secret)
   |
   I6 +--> DE6: regressions caught automatically; ROP laws confirmed by tests
   |
   I7 +--> DE7: launch reproducible with one command; UI covers scenarios
   |
   I8 +--> DE8: guest CRUD scenarios (creation, list, view, search,
   |            editing, deletion) implemented as use-cases on top of
   |            I1..I5 and available via HTTP handlers

                      ┌──────────────────────────────────────────────────┐
  DE1..DE8 +--> DE0 │ DE0 (desired effect): working, safe and           │
  (DE0 because    │ verifiable MVP of guest accounting system — all DE │
   DE1..DE8 needed │ are necessary and together sufficient for acceptance)│
   and together  └──────────────────────────────────────────────────┘
   sufficient)
```

### 3.2 Negative Branch Verification

```
NB-1: I1 + "simple use-cases" +--> NB-1: ROP adds noise to trivial operations
      +--> Remed: Fold executed at handler boundary (transport folds Either into
                   HTTP response); for trivial use-cases exactly one Fold point
                   per handler is allowed, service chains ≤ 3 operations; domain
                   and service do not "leak" into UI (REQ-API-002).
                   BRANCH CLOSED: noise does not reach the user and does not
                   lengthen the path of new endpoints.

NB-2: I4(index user_id, created_at) +--> NB-2: index slightly slows INSERT into guests
      +--> Remed: at MVP volumes insert cost is negligible; plan correctness
                   confirmed by EXPLAIN ANALYZE (REQ-DB-003, GATE-12);
                   at data growth pg_trgm/GIN considered for search.
                   BRANCH CLOSED: costs measured and acceptable.

NB-3: I4(FK guests.user_id → users.id, cascade) +--> NB-3: user deletion
      cascadingly deletes their guests
      +--> Remed: conscious design decision per 05-database.md §3.1, fixed by
                   REQ-DB-002 constraints; alternative (RESTRICT) conflicts with
                   accepted accounting model. BRANCH CLOSED: behavior documented
                   and covered by TST-DB-003..005.

NB-4: I3 (JWT with lifetime) +--> NB-4: expired token "logs out" user
      mid-scenario
      +--> Remed: client on 401 resets token and redirects to login
                   (REQ-FE-002, REQ-AUTH-003) — scenario recoverable, secret
                   not compromised. BRANCH CLOSED.

NB-5: I7 + REQ-FE-002 (JWT stored on client) +--> NB-5: XSS vector —
      token can be extracted by script from page
      +--> Remed: FE contour — internal tool without user-generated content;
                   token stored in tab memory (not localStorage), CSP forbids
                   third-party scripts; httpOnly-cookie unavailable in declared
                   single-domain contour. BRANCH CLOSED with fixed limitation:
                   upon user-content appearance mandatory transition to
                   httpOnly-cookie, fixed by REQ-FE-002 revision.
```

### 3.3 Positive Stimuli (Lifters)

```
P1: I1 + I6 +--> P1: ROP law tests (TST-ROP-001..007) become living
       |            documentation of the contract; new developers read tests,
       |            not wiki → project onboarding speeds up (supports DE6).
       |            Criterion: new developer builds project and runs
       |            TST-ROP-001..007 within 1 working day (onboarding log).
       |
       +--> P2: I7 +--> P2: one command docker compose up closes both
                  development and acceptance GATE-10..12 → smoke check becomes
                  routine (strengthens DE7). Criterion: smoke `docker compose up` +
                  GATE-10 run before each acceptance decision —
                  no less than once per sprint (run log).
```

### 3.4 Explanation

FRT shows that a safe MVP is not "CRUD + security later", but **a set of eight
one-time injections**, each eliminating a root from CRT: I1 heals U5; U2 healed by
CLR NE-2 stage (go.mod, directory skeleton) — an injection preceding FRT; I3 together
with U10 closes U6; I4 — U3; I2 heals U11, and composition of failures U7 closed by
I1+I2+I5; I6 — U8; I7 — U9; I8 closes U4 (contract and layers), and I8+I2+I5+I6
together — U1. Negative branches NB-1..NB-5 checked: all closed by remediations
recorded in contracts (04, 05, 06) and criteria (08), or contain a fixed limitation
(NB-5). Positive stimuli P1, P2 have measurable criteria and show injections pay off
wider than the original goal (documentation, acceptance acceleration).

### 3.5 Connection to Requirements

| FRT node | REQ |
|---|---|
| I1/DE1 | REQ-ROP-001, REQ-ROP-002, REQ-ROP-003 |
| I2/DE2 | REQ-GUEST-001, REQ-GUEST-005, REQ-API-003 |
| I3/DE3 | REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003, REQ-GUEST-007, REQ-SEC-001, REQ-SEC-002 |
| I4/DE4 | REQ-DB-001, REQ-DB-002, REQ-DB-003, REQ-DB-004 |
| I5/DE5 | REQ-API-001, REQ-API-002, REQ-SEC-003 |
| I6/DE6 | REQ-INFRA-004 |
| I7/DE7 | REQ-INFRA-001, REQ-INFRA-002, REQ-INFRA-003, REQ-FE-001, REQ-FE-002 |
| I8/DE8 | REQ-GUEST-001..006, REQ-API-001 |
| NB-1..NB-5 | REQ-API-002, REQ-DB-002, REQ-DB-003, REQ-AUTH-003, REQ-FE-002 |
| DE0 | REQ-GUEST-001..006 (CRUD within isolation), REQ-API-001 |

---

## 4. PRT — Prerequisite Tree

Goal G0: achieve FRT — working, safe, verifiable MVP. Each obstacle OB-n is
overcome by a necessary condition (injection) NE-n bound to a life cycle stage
(SDD, CLR, TDD L1, Stubs, GATE-1, Translation, Security-Core, TDD L2, Refactor,
Security/Infra, GATE-3).

### 4.1 Scheme

```
                          ┌────────────────────────────────────────────────┐
                          │ G0 (goal): FRT achieved — MVP of guest       │
                          │ accounting system accepted (DE0 from section 3)│
                          └───────────────────────▲────────────────────────┘
                                                  │ +--> G0
        (overcoming all obstacles NE-1..NE-11 sequentially leads to the goal)

  OB-1: specs approved, but not yet linked   │ NE-1: keep 01..08 as source of truth
  to code and tests (GATE-1 not              │ at all subsequent stages; do not diverge
   run)                                      │ from REQ at CLR and development
   +--> [stage SDD] ────────────────────────┤ NE-1 because OB-1: without an immutable source
                                            │ of truth each stage makes decisions "from memory"
                                            ▼
  OB-2: no Go project structure              │ NE-2: go.mod + directory skeleton per 03 §10;
  (empty directory)                          │ layer layout review, port/adapter
   +--> [stage CLR] ────────────────────────┤ NE-2 because OB-2: nowhere to write; CLR fixes
                                            │ layer boundaries before the first file; GATE-2 check
                                            ▼
  OB-3: ROP invariants not verified          │ NE-3: TDD L1: red→green cycles by
                                            │ TST-ROP-001..014 (Either, IOEither) + domain
   +--> [stage TDD L1] ────────────────────┤ NE-3 because OB-3: without proof of ROP laws
                                            │ injection I1 from FRT — declaration, not fact
                                            ▼
  OB-4: service layer cannot be tested     │ NE-4: Stubs: in-memory repository adapters
  without DB                               │ (port/adapter); SQL migrations moved to NE-7
   +--> [stage Stubs] ─────────────────────┤ NE-4 because OB-4: stubs separate domain/service
                                            │ from infrastructure
                                            ▼
  OB-5: REQ↔Test traceability not verified │ NE-5: audit of matrix 01↔07: each REQ ≥ 1 Test ID,
  in practice                               │ no "hanging" Test ID
   +--> [stage GATE-1] ────────────────────┤ NE-5 because OB-5: without the gateway acceptance is subjective
                                            ▼
  OB-6: contract exists only in one locale │ NE-6: Translation: create and synchronize
                                           │ specs/en/ based on base frozen by GATE-1,
   +--> [stage Translation] ───────────────┤ preserving REQ/Test/GATE IDs
                                           │ NE-6 because OB-6: after the gateway IDs are frozen —
                                           │ translation synchronizes the frozen base without
                                           │ creating divergent spec versions
                                            ▼
  OB-7: auth, schema, handlers not        │ NE-7: Security-Core: Argon2id + JWT HS256 +
  implemented                              │ middleware + user_id filter (I3), SQL migrations
   +--> [stage Security-Core] ────────────┤ users/guests with FK/CHECK/index (I4), unified
                                           │ error format (I5), use-cases + handlers (I8)
                                           │ NE-7 because OB-7: TDD L2 (NE-8) proves only
                                           │ implemented — auth, schema, handlers must exist
                                           │ before integration tests
                                            ▼
  OB-8: HTTP behavior and DB not covered │ NE-8: TDD L2: httptest tests (auth, guests, api)
                                           │ + integration with isolated PostgreSQL
   +--> [stage TDD L2] ───────────────────┤ NE-8 because OB-8: isolation and statuses proved
                                           │ only by a run through the real contour
                                            ▼
  OB-9: layers grow, duplicate each other  │ NE-9: Refactor: refactoring on green tests,
                                           │ without behavior change (tests are insurance)
   +--> [stage Refactor] ─────────────────┤ NE-9 because OB-9: accumulating copy-paste
                                           │ breaks further acceleration
                                            ▼
  OB-10: no reproducible launch and FE   │ NE-10: Security/Infra: Docker Compose,
                                           │ .env.example, README, frontend TS
   +--> [stage Security/Infra] ───────────┤ NE-10 because OB-10: acceptance GATE-10..12
                                           │ requires one launch command
                                            ▼
  OB-11: readiness not provable          │ NE-11: GATE-3: full go test ./... run +
                                           │ tsc, smoke docker compose up (GATE-4..12)
   +--> [stage GATE-3] ──────────────────┤ NE-11 because OB-11: without full gateway
                                           │ DE0 cannot be claimed
                                            ▼
                          ┌────────────────────────────────────────────────┐
                          │ G0 achieved: DE0 confirmed by GATE-1..12       │
                          └────────────────────────────────────────────────┘
```

### 4.2 Links (A --> B (B because A))

For each pair the specific NE-n property is indicated, without which the next
obstacle is insurmountable.

```
OB-1  --> NE-1  (NE-1 because OB-1: without immutable source of truth decisions are
                 taken from memory (= D1 CRT))
NE-1  --> OB-2  (OB-2 because NE-1: layout inherits layers from approved 03 §10 —
                 without immutable spec there is no criterion "layout is correct")
NE-2  --> OB-3  (OB-3 because NE-2: ROP types written as internal/rop package inside
                 fixed build unit — without go.mod and directories go test does not build)
NE-3  --> OB-4  (OB-4 because NE-3: Either/IOEither ready — repository ports typed
                 Either[E,T], service composed on stubs; without ready types stubs are
                 incompatible with service)
NE-4  --> OB-5  (OB-5 because NE-4: code and tests exist — traceability matrix can be
                 checked against fact, not plan)
NE-5  --> OB-6  (OB-6 because NE-5: after the gateway REQ/Test/GATE IDs are frozen —
                 translation synchronizes the frozen base and does not create diverging spec versions)
NE-6  --> OB-7  (OB-7 because NE-6: route names and error codes identical in both locales —
                 auth/handlers implemented once, no rewrite after translation)
NE-7  --> OB-8  (OB-8 because NE-7: auth, migrations, error format and handlers exist —
                 TDD L2 proves implemented; tests against unrealized contour inapplicable)
NE-8  --> OB-9  (OB-9 because NE-8: on green L2 refactoring can be done safely)
NE-9  --> OB-10 (OB-10 because NE-9: clean structure — Compose/README/FE embedded without sorting copy-paste)
NE-10 --> OB-11 (OB-11 because NE-10: everything implemented — only readiness proof by full run remains)
NE-11 --> G0    (G0 because NE-11: GATE-3 green + GATE-1..12 → DE0 achieved)
```

### 4.3 Explanation

PRT — the reverse side of CRT: to each UDE corresponds an obstacle, to each
obstacle — exactly one injection and one stage. The order NE-1..NE-11 reflects the
critical chain "spec → structure → unit proofs → stubs → traceability → translation →
auth/schema implementation → integration → refactoring → infrastructure/FE → readiness gateway".
Two order changes are fundamental:

1. **Translation** stands after GATE-1: translation synchronizes the already
   frozen and traced base, not a moving spec.
2. **Security-Core** (Argon2id, JWT, middleware, migrations, handlers) stands
   before TDD L2: proof (tests) cannot precede implementation.

Stage SDD is already passed (TASK-002: `01..08` approved) — NE-1 = keep spec as source
of truth and do not diverge from it at CLR and development. SQL migrations moved
out of NE-4 (Stubs) into NE-7, because they are not needed for testing the service
layer on stubs, but are necessary for integration tests.

### 4.4 Connection to Requirements

NE-1: all REQ-*. NE-2: REQ-ROP-001, REQ-ROP-002 (architectural boundaries 03).
NE-3: REQ-ROP-003, REQ-GUEST-001, REQ-GUEST-005 (domain in TDD L1). NE-4:
REQ-ROP-002. NE-5: REQ-INFRA-004 (traceability 07). NE-6: REQ-API-001..003.
NE-7: REQ-AUTH-001..003, REQ-SEC-001..003, REQ-DB-001..004, REQ-GUEST-007,
REQ-API-002, REQ-GUEST-001..006. NE-8: REQ-AUTH-001..003, REQ-GUEST-001..007,
REQ-DB-002..004, REQ-API-001..003, REQ-SEC-003. NE-9: REQ-INFRA-004.
NE-10: REQ-INFRA-001..003, REQ-FE-001, REQ-FE-002.
NE-11: REQ-INFRA-004 (GATE-3).

---

## 5. Transition Tree — Transition Tree

Sequence of steps from CRT to FRT. Step format: "If [action], then [effect],
because [justification/removed UDE], verified by [REQ/Test/GATE]".

### 5.1 Scheme

```
 CRT (D1, U2..U11)
   │
   │  STEP 1 [SDD]
   │  If approve 01..08 as source of truth and do not change them silently,
   │  then each function gets a verifiable "done" definition,
   │  because D1 requires a verifiable link "requirement → code → test";
   │  verified by: composition 01..08, 29 REQ, GATE-1 (matrix REQ ↔ Test ↔ AC).
   +-->
   │  STEP 2 [CLR]
   │  If create go.mod and directory skeleton per 03 §10 (cmd, internal/{config,
   │  domain, rop, service, repository/postgres, transport}) with layout review,
   │  then a build unit and layer boundaries appear, because U2 (no structure)
   │  blocks everything else; verified by: go build ./..., go vet (GATE-2),
   │  compliance with 03 §2, §10.
   +-->
   │  STEP 3 [TDD L1]
   │  If first write tests TST-ROP-001..014, then implement Either/IOEither
   │  until green, then ROP invariants are proven (I1/DE1), because U5 (no ROP)
   │  generates U6 and U7 through error loss; verified by:
   │  REQ-ROP-001..003, TST-ROP-001..014, GATE-6.
   +-->
   │  STEP 4 [TDD L1, continuation]
   │  If describe User/Guest domain via factories (validation of full name,
   │  dates, check_out > check_in) with red→green tests, then invalid input
   │  is filtered before DB (I2/DE2), because U7/U11 (no validation and factories);
   │  verified by: REQ-GUEST-001,005, REQ-API-003, TST-GUEST-002,003.
   +-->
   │  STEP 5 [Stubs]
   │  If implement repository ports and in-memory stubs, then service layer
   │  on top of them (Either pipeline decode→validate→insert), then logic
   │  tested without DB, because OB-4; verified by: service unit tests,
   │  REQ-ROP-002 (Run once). SQL migrations not needed here — they are in Step 8.
   +-->
   │  STEP 6 [GATE-1]
   │  If verify matrix 01↔07 (each REQ ≥ 1 Test ID, no hanging Test ID)
   │  and fix discrepancies before writing HTTP tests, then acceptance becomes
   │  objective, because OB-5; verified by: GATE-1, REQ-INFRA-004.
   +-->
   │  STEP 7 [Translation]
   │  If create and synchronize specs/en/ based on base frozen by GATE-1,
   │  preserving REQ/Test/GATE IDs, then contract is single for all participants
   │  and subsequent HTTP tests written by unchanged names, because OB-6;
   │  verified by: ru/en section parity, ID invariance.
   +-->
   │  STEP 8 [Security-Core]
   │  If implement Argon2id + JWT HS256 (secret from env) with middleware
   │  and user_id filter (I3), versioned SQL migrations users/guests
   │  with FK/CHECK/index (I4), unified error format at boundary (I5)
   │  and connect GuestService/AuthService use-cases with HTTP handlers (I8),
   │  then auth contour, schema and handlers exist before integration tests,
   │  because implementation must precede proof (OB-7);
   │  verified by: REQ-AUTH-001..003, REQ-SEC-001,002, REQ-DB-001,002,
   │  REQ-GUEST-001..007, REQ-API-001,002.
   +-->
   │  STEP 9 [TDD L2]
   │  If write httptest tests (auth, guests, api) and integration with
   │  isolated PostgreSQL, including user isolation and transactions,
   │  then boundary and DB behavior proven (I3..I6, I8), because U6/U8;
   │  verified by: REQ-AUTH-001..003, REQ-GUEST-001..007, REQ-DB-002..004,
   │  REQ-API-001..003, REQ-SEC-003, TST-*, GATE-4, GATE-7, GATE-8, GATE-12.
   +-->
   │  STEP 10 [Refactor]
   │  If on fully green tests refactor layer duplication and Either composition
   │  chains without changing observable behavior, then code remains maintainable,
   │  because OB-9; verified by: invariance of green go test ./... run before/after
   │  (duplicate reduction — by lint).
   +-->
   │  STEP 11 [Security/Infra]
   │  If implement Docker Compose, .env.example, README and frontend TS,
   │  then U9 and FRT injection I7 closed, because reproducibility
   │  and FE are part of MVP, not "later"; verified by: REQ-INFRA-001..003,
   │  REQ-FE-001,002, GATE-5, GATE-10, GATE-11.
   +-->
   │  STEP 12 [GATE-3]
   │  If run go test ./... (all domains), tsc and smoke compose stack,
   │  then system is ready for acceptance (DE0), because OB-11; verified by:
   │  GATE-3, GATE-4..GATE-12, all AC from 08 §1.
   +-->
  FRT (DE0 achieved; CRT-UDE eliminated)
```

### 5.2 Links (A --> B (B because A))

```
CRT (D1) --> STEP 1 (STEP 1 because D1: verifiable foundation starts with spec)
STEP 1 --> STEP 2 (STEP 2 because STEP 1: structure inherits layers from approved 03)
STEP 2 --> STEP 3 (STEP 3 because STEP 2: ROP types build as a package in go.mod)
STEP 3 --> STEP 4 (STEP 4 because STEP 3: domain built on already proven Either/IOEither)
STEP 4 --> STEP 5 (STEP 5 because STEP 4: service layer composes domain and ports)
STEP 5 --> STEP 6 (STEP 6 because STEP 5: there is something to trace in REQ↔Test matrix)
STEP 6 --> STEP 7 (STEP 7 because STEP 6: after GATE-1 IDs are frozen — translation does not diverge)
STEP 7 --> STEP 8 (STEP 8 because STEP 7: routes and error codes identical before implementation)
STEP 8 --> STEP 9 (STEP 9 because STEP 8: auth, schema and handlers exist — they can be proven)
STEP 9 --> STEP 10 (STEP 10 because STEP 9: green integration tests insure refactoring)
STEP 10 --> STEP 11 (STEP 11 because STEP 10: clean structure accommodates infra/FE without copy-paste)
STEP 11 --> STEP 12 (STEP 12 because STEP 11: everything implemented — only full proof remains)
STEP 12 --> FRT (FRT because STEP 12: GATE-3 green ⇒ DE0 ⇒ negative branches closed)
```

### 5.3 Explanation

Each step eliminates exactly one CRT cause and relies on the previous result
(PRT chain NE-1..NE-11). TDD steps are split into L1 (proof of types and domain
without DB) and L2 (proof of boundary and DB behavior), and Stubs separate
infrastructure from logic so L1 cycles do not require PostgreSQL.
Translation stands after GATE-1 — translation synchronizes the frozen base.
**Auth implementation (Argon2id/JWT/middleware) and SQL migrations moved to Step 8
[Security-Core] and stand before TDD L2**: tests cannot prove what does not yet exist.
Refactor strictly between green L2 and Security/Infra: refactoring without tests is
forbidden, infrastructure without refactoring would copy-paste layers.
GATE-1 before L2 — so as not to write tests "into nowhere"; GATE-3 last — as the
final gateway.

### 5.4 Connection to Requirements

Step 1: all REQ-*. Step 2: REQ-ROP-001..003 (layer boundaries from 03), GATE-2. Step 3:
REQ-ROP-001..003. Step 4: REQ-GUEST-001,005, REQ-API-003. Step 5: REQ-ROP-002.
Step 6: REQ-INFRA-004. Step 7: REQ-API-001..003. Step 8: REQ-AUTH-001..003,
REQ-SEC-001,002, REQ-DB-001,002, REQ-GUEST-001..007, REQ-API-001,002. Step 9:
REQ-AUTH-001..003, REQ-GUEST-001..007, REQ-DB-003,004, REQ-API-001..003,
REQ-SEC-003. Step 10: REQ-INFRA-004. Step 11: REQ-INFRA-001..003, REQ-FE-001,002.
Step 12: REQ-INFRA-004, REQ-DB-003 (GATE-12).

---

## 6. Strategy & Tactics Tree — Strategy and Tactics Tree

### 6.1 Scheme

```
  VISION: safe, type-safe and verifiable guest accounting system,
          built on approved specification (SDD)
                     │
                     ▼
 ┌──────────────────────────────────────────────────────────────────────────────┐
 │ NECESSARY ASSUMPTIONS (N-nodes):                                             │
 │  NR-1: acceptance requires reproducibility and provability of result         │
 │  NR-2: guest data (passwords, someone else's records) cannot be trusted      │
 │          "approximately"                                                     │
 └───────────────────────────┬──────────────────────────────────────────────────┘
                             │ NR-1, NR-2 --> S1
                             ▼
 ┌──────────────────────────────────────────────────────────────────────────────┐
 │ STRATEGY S1: specification — single source of truth; build "bottom-up" —     │
 │ first foundation (types, domain, schema, contract, tests), then functions    │
 │ on top of it; each function inherits safety and error handling from foundation │
 │ (Cloud conflict removal: strictness = speed)                                 │
 └───────────────────────────┬──────────────────────────────────────────────────┘
                             │
    ┌────────────────────────┼─────────────────────────────────────┐
    │                        │                                     │
    ▼                        ▼                                     ▼
  TACTIC T1                TACTIC T2                            TACTIC T3
  ROP as the only          domain — pure functions               security by default
  composition method                       │                         │
    │ T1.1 Either[E,T] for     │ T2.1 NewUser/NewGuest factories │ T3.1 user_id only from
    │ synchronous steps        │ return Either                   │ token
    │ T1.2 IOEither for        │ T2.2 date/full name validation  │ T3.2 user_id filter in
    │ effects, Run once        │ before DB                       │ all SQL queries
    │ T1.3 (T, error) only   │ T2.3 DomainError without panic  │ T3.3 someone else's record = 404
    │ at adapter boundaries  │                                     │ T3.4 Argon2id, secret from env
    │ Stage: TDD L1          │ Stage: TDD L1                     │ Stage: Security-Core,
    │ Criterion:             │ Criterion:                        │ TDD L2
    │ TST-ROP-001..014,    │ TST-GUEST-002,003; domain tests   │ Criterion: TST-AUTH-*,
    │ GATE-6               │ REQ-GUEST-001,005, REQ-API-003  │ TST-SEC-001,002,
    │ REQ-ROP-001..003     │                                     │ TST-GUEST-009..015
    │                        │                                     │ REQ-AUTH-001..003,
    │                        │                                     │ REQ-SEC-001,002,
    │                        │                                     │ REQ-GUEST-007
    │                        │                                     │
    ▼                        ▼                                     ▼
  TACTIC T4                TACTIC T5                            TACTIC T6
  integrity — safety nets  reproducible                         feedback via automation
  at DB level              launch                               │
    │ T4.1 FK + NOT NULL +   │ T5.1 Docker Compose +             │ T6.1 unit tests (ROP,
    │ CHECK (check_out       │ .env.example                      │ domain) without DB
    │ > check_in)            │ T5.2 README with executable       │ T6.2 httptest tests
    │ T4.2 index             │ commands                          │ T6.3 integration with
    │ (user_id, created_at)  │ T5.3 FE TS compiles tsc           │ isolated DB
    │ T4.3 migrations        │                                   │ T6.4 gateways: GATE-1,
    │ versioned and          │ Stage: Security/Infra             │ GATE-3 (+4..12)
    │ idempotent             │ Criterion: TST-INFRA-001..003,  │ Stage: TDD L1/L2,
    │ Stage: Security-Core,  │ TST-FE-001..003, GATE-10, GATE-11 │ GATE-1, GATE-3
    │ Criterion: TST-DB-001..│ REQ-INFRA-001..003,               │ REQ-INFRA-004
    │ 007, GATE-9, GATE-12   │ REQ-FE-001,002                    │
    │ REQ-DB-001..004        │                                   │
```

### 6.2 Consistency and Required Assumptions

Links (A --> B (B because A)) between levels:

```
Vision --> NR-1 (NR-1 because Vision: acceptance without reproducibility
                 and provability is unattainable)
Vision --> NR-2 (NR-2 because Vision: "safe system" requires mandatory
                 password and someone else's record protection)
NR-1 --> S1 (S1 because NR-1: "bottom-up" strategy builds foundation
             which can be reproduced and accepted)
NR-2 --> S1 (S1 because NR-2: security embedded into foundation, not added later)
S1 --> T1 (T1 because S1: strictness starts with types — ROP foundation)
S1 --> T2 (T2 because S1: domain inherits Either from T1)
S1 --> T3 (T3 because S1: isolation is part of foundation, not "setting")
S1 --> T4 (T4 because S1: schema is versioned part of foundation)
S1 --> T5 (T5 because S1: acceptance requires reproducible launch)
S1 --> T6 (T6 because S1: without automatic feedback strictness would
           degenerate into bureaucracy — T6 proves S1 works)
T1 --> T2 (T2 because T1: domain inherits Either and DomainError from T1)
T1 --> T6 (T6 because T1: test errors are Left[E] values, not exceptions)
```

T4 and T5 rely on S1 directly: DB schema and reproducible launch do not depend
on ROP types T1. The false universal link "T1 --> T2..T6" eliminated.

Required assumptions (verifiable):

| ID | Assumption | How verified |
|---|---|---|
| A1 | Strictness at start reduces total time (no rework) | GATE-1 passed before L2 tests; time to implement use-case #2 ≤ use-case #1 (commit log); regressions after Refactor = 0 |
| A2 | ROP investment pays off on second and subsequent use-cases | time to implement each next use-case ≤ first (commit log); handlers have no manual error handling (review) |
| A3 | `user_id` filter in repository excludes leaks at correct implementation | GATE-7: TST-GUEST-009,011,013,015 |
| A4 | IOEither does not replace timeouts/transactions, but improves composition | applicability boundary fixed in REQ-ROP-003 (note); transactions — REQ-DB-004 |

### 6.3 Explanation

Strategy S1 — direct Evaporating Cloud resolution: "fast" achieved not by cutting
strictness, but by its **one-time installation** into the foundation. N-nodes
NR-1/NR-2 fix why Vision cannot be achieved without S1. Tactics T1..T6 do not
conflict: T1 and T2 work on TDD L1 (without DB), T3 and T4 close isolation and
integrity on Security-Core/TDD L2/Stubs, T5 closes reproducibility (Security/Infra),
T6 closes the feedback loop via GATE gateways. Each tactic has a measurable criterion
(Test ID / GATE), linking TOC analysis with `07-test-plan.md` and `08-acceptance-criteria.md`.

### 6.4 Connection to Requirements

T1: REQ-ROP-001..003. T2: REQ-GUEST-001,005, REQ-API-003. T3: REQ-AUTH-001..003,
REQ-SEC-001,002, REQ-GUEST-007. T4: REQ-DB-001..004. T5: REQ-INFRA-001..003,
REQ-FE-001,002. T6: REQ-INFRA-004 (and all REQ via GATE traceability).

---

## 7. Summary Table: TOC Tool → Stage → REQ

| TOC tool | Key finding | Stage(s) | REQ |
|---|---|---|---|
| CRT | root D1 — missing versioned, testable foundation; U2..U11 — consequences, U1 — final UDE | SDD | all REQ-* |
| Evaporating Cloud | "strictness = speed" with one-time foundation installation; assumptions "fast without types" and "strict = slow" destroyed | SDD, TDD L1 | REQ-ROP-001..003, REQ-AUTH-001..003, REQ-GUEST-001..007, REQ-API-001..003, REQ-DB-001..004, REQ-INFRA-004 |
| FRT (I1..I8, NB-1..NB-5, P1..P2) | 8 injections → DE0; negative branches closed (Fold, EXPLAIN, cascade, 401→login, JWT storage limitation) | all stages | REQ-ROP-001..003, REQ-GUEST-001..007, REQ-AUTH-001..003, REQ-DB-001..004, REQ-SEC-001..003, REQ-API-001..003, REQ-INFRA-001..004, REQ-FE-001,002 |
| PRT (OB-1..11 / NE-1..11) | sequential overcoming chain bound to SDD→GATE-3; Security-Core before TDD L2 | SDD..GATE-3 | all REQ-* (NE per domains — see 4.4) |
| Transition Tree (steps 1..12) | "if → then → because → verified" at each stage; auth implementation before its tests | SDD → CLR → TDD L1 → Stubs → GATE-1 → Translation → Security-Core → TDD L2 → Refactor → Security/Infra → GATE-3 | all REQ-* |
| Strategy & Tactics (S1, NR-1..NR-2, T1..T6, A1..A4) | strategy "spec → foundation → functions" with measurable tactics | TDD L1/L2, Stubs, Security-Core, Security/Infra, GATE | all REQ-* |

**REQ used (all 29):** REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003,
REQ-GUEST-001, REQ-GUEST-002, REQ-GUEST-003, REQ-GUEST-004, REQ-GUEST-005,
REQ-GUEST-006, REQ-GUEST-007, REQ-ROP-001, REQ-ROP-002, REQ-ROP-003,
REQ-DB-001, REQ-DB-002, REQ-DB-003, REQ-DB-004, REQ-SEC-001, REQ-SEC-002,
REQ-SEC-003, REQ-API-001, REQ-API-002, REQ-API-003, REQ-FE-001, REQ-FE-002,
REQ-INFRA-001, REQ-INFRA-002, REQ-INFRA-003, REQ-INFRA-004.
