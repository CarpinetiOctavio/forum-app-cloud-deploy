# Testing — Existing Suite and Rules

## Purpose

This file documents the existing test suite inherited from forum-app-qa-pipeline and defines the rules for testing in this project. Read `docs/rules/constraints.md` before this file. Any conflict resolves in favor of `constraints.md`.

---

## Existing test suite — do not modify

The test suite was established in forum-app-qa-pipeline and must be preserved exactly as inherited. Do not rename, restructure, or delete any existing test. Do not lower coverage thresholds.

| Layer | Tool | Tests | Coverage |
|-------|------|-------|----------|
| Backend unit tests | Go testing | 35 tests | 86.5% |
| Frontend unit tests | Jest | 39 tests | 92.44% |
| E2E tests | Cypress | 15 tests | — |
| Static analysis | SonarCloud | — | Quality Gate: PASSED |
| **Total** | | **89 tests** | |

---

## Running tests locally

**Backend unit tests:**
```bash
cd backend

# Run all tests
go test ./tests/services/... -v

# Run with coverage
go test ./tests/services/... -v -cover -coverpkg=./internal/services/...

# Generate HTML coverage report
go test ./tests/services/... -coverprofile=coverage.out -coverpkg=./internal/services/...
go tool cover -html=coverage.out

# Coverage summary in terminal
go tool cover -func=coverage.out
```

**Frontend unit tests:**
```bash
cd frontend

# Run once
npm test -- --watchAll=false

# Run with coverage
npm test -- --coverage --watchAll=false

# Open HTML coverage report
open coverage/lcov-report/index.html
```

**E2E tests — Cypress:**
```bash
# Backend and frontend must be running before executing Cypress

# Terminal 1
cd backend && go run cmd/api/main.go

# Terminal 2
cd frontend && npm start

# Terminal 3
cd frontend && npx cypress run        # headless (CI/CD)
cd frontend && npx cypress open       # interactive (local)
```

---

## Coverage gates

These thresholds are enforced in the pipeline and must not be lowered:

| Layer | Minimum | Current |
|-------|---------|---------|
| Backend | 70% | 86.5% |
| Frontend | 70% | 92.44% |

---

## Test structure — backend

```
backend/tests/
├── services/
│   ├── auth_service_test.go     # Authentication logic tests
│   └── post_service_test.go     # Post and comment logic tests
└── mocks/
    ├── mock_user_repository.go  # Mock for user data access
    └── mock_post_repository.go  # Mock for post data access
```

Tests cover `internal/services/` exclusively. Handlers and repository implementations are not unit tested — they are covered by E2E tests via Cypress.

---

## Test structure — frontend

```
frontend/src/
├── components/
│   ├── Login/Login.test.tsx
│   ├── PostList/PostList.test.tsx
│   └── [other components]/*.test.tsx
└── services/
    ├── authService.test.ts
    └── postService.test.ts
```

---

## Test structure — Cypress

```
frontend/cypress/e2e/blog/
├── auth.cy.js        # 5 tests — registration and login flows
├── posts.cy.js       # 5 tests — post creation, listing, deletion
├── comments.cy.js    # 4 tests — comment creation and listing
└── full-flow.cy.js   # 1 test  — complete user journey
```

---

## Naming convention for new tests

Existing tests retain their current naming. Any new test added in this project must follow the `Should_When` convention:

```go
// Go — backend
func TestRegister_WhenEmailAlreadyExists_ShouldReturnError(t *testing.T) {}
func TestCreatePost_WhenUserIsNotAuthenticated_ShouldReturnUnauthorized(t *testing.T) {}
```

```typescript
// TypeScript — frontend
it('should return error when email already exists', () => {})
it('should render post list when user is authenticated', () => {})
```

Internal test structure follows AAA (Arrange, Act, Assert):
- `sut` — System Under Test
- `got` — actual result
- `want` / `expected` — expected result

---

## What Code must do before adding or modifying any test

1. Confirm that all 89 existing tests pass locally before making any change.
2. Never modify an existing test to make it pass — fix the source code instead.
3. If a new test is required, follow the `Should_When` naming convention and AAA structure.
4. Report current test status before making any pipeline or Dockerfile change that could affect test execution.