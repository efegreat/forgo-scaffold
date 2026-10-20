# Testing

This directory contains tests that exercise the application beyond isolated unit tests.

The general rule is:

- Unit tests stay next to the code they test.
- Tests under `/test` may use real infrastructure such as MySQL, PostgreSQL, Redis, RabbitMQ, or other external dependencies.
- Infrastructure should preferably be disposable and reproducible, for example through Testcontainers.
- Tests should remain ordinary Go tests. Avoid introducing Gherkin-style abstractions unless there is a strong reason to do so.

```text
test/
├── fixtures/
├── http/
│   ├── features/
│   └── e2e/
└── integration/
```

## `fixtures/`

Contains reusable test data used to establish an application's initial state.

SQL fixtures are preferred for relational databases because they are explicit, easy to inspect, and support database-specific behavior without introducing another data representation.

Example:

```text
test/
└── fixtures/
    └── users/
        ├── empty.sql
        ├── active_user.sql
        └── suspended_user.sql
```

A fixture should describe a useful test state rather than mirror individual database tables.

Prefer:

```text
active_user.sql
user_with_orders.sql
expired_session.sql
```

over:

```text
users.sql
orders.sql
sessions.sql
```

when the fixture exists specifically to prepare a business scenario.

Example:

```sql
-- fixtures/users/active_user.sql

INSERT INTO users (
    id,
    email,
    status
) VALUES (
    1,
    'user@example.com',
    'active'
);
```

Fixtures may be shared between feature, E2E, and integration tests.

Tests are responsible for ensuring that the database is reset into a predictable state before applying fixtures.

---

# HTTP tests

Tests under `http/` exercise the application through its HTTP boundary.

They should use the real HTTP configuration whenever practical, including:

- routing
- middleware
- request binding
- validation
- handlers
- serialization
- error mapping

Application dependencies used by the request path may be backed by real test infrastructure.

HTTP tests are divided into **features** and **E2E** tests based on the scope of behavior being tested.

## `http/features/`

Feature tests verify one HTTP endpoint or one closely related HTTP behavior.

A feature test normally checks:

1. the initial state,
2. one HTTP request,
3. the HTTP response,
4. relevant side effects.

For example:

```text
Given an active user exists

POST /v1/users/1/settings

Expect:
- HTTP 200
- the returned settings are correct
- the settings are persisted in the database
```

The Go test should remain direct and readable:

```go
func TestUpdateSettings(t *testing.T) {
    env := support.NewHTTP(t)

    env.Fixture(t, "users/active_user.sql")

    resp := env.Post(t, "/v1/users/1/settings", UpdateSettingsRequest{
        Language: "en",
    })

    resp.Status(t, http.StatusOK)

    env.DBHas(t, "user_settings", map[string]any{
        "user_id":  1,
        "language": "en",
    })
}
```

A feature test may use a real database, cache, broker, or other infrastructure when those dependencies are part of the endpoint's behavior.

A feature test should generally stay focused on a single endpoint.

Good examples:

```text
create_user_test.go
login_test.go
update_profile_test.go
cancel_order_test.go
```

Avoid turning one feature test into a complete business journey involving many unrelated endpoints. That belongs under `e2e/`.

---

## `http/e2e/`

E2E tests exercise a business flow through multiple HTTP requests.

They verify that independently exposed HTTP operations work correctly when used together in the way a real client or business process would use them.

For example:

```text
register user
    ↓
login
    ↓
update profile
    ↓
retrieve profile
```

A test could look like:

```go
func TestUserOnboarding(t *testing.T) {
    env := support.NewHTTP(t)

    user := register(t, env, "user@example.com", "secret")
    token := login(t, env, user.Email, "secret")

    updateProfile(t, env, token, UpdateProfileRequest{
        DisplayName: "John",
    })

    profile := getProfile(t, env, token)

    require.Equal(t, "John", profile.DisplayName)
}
```

Helpers such as `register`, `login`, and `getProfile` should still perform real HTTP requests.

Their purpose is to remove repetitive HTTP plumbing so the test itself reads like the business flow being verified.

E2E tests should not bypass HTTP by directly calling services for intermediate steps.

If the test begins as an HTTP E2E test, the important business interactions in that scenario should continue through the HTTP boundary.

E2E tests do **not** necessarily mean testing the complete production deployment.

They may run the real ForGo HTTP handler through a local test server while using real backing infrastructure.

The purpose is to verify the externally observable HTTP workflow, not Kubernetes, ingress, DNS, or production networking.

---

# `integration/`

Integration tests verify a particular component together with real infrastructure.

Unlike HTTP tests, they do not need to exercise the HTTP server.

Typical targets include:

- repositories against a real database,
- services using a real database or Redis instance,
- message handlers using database state,
- outbox or inbox storage implementations,
- infrastructure adapters,
- commands that depend on real infrastructure.

Example repository test:

```go
func TestUserRepository_Create(t *testing.T) {
    db := support.NewDatabase(t)

    repo := NewUserRepository(db)

    user, err := repo.Create(t.Context(), User{
        Email: "user@example.com",
    })

    require.NoError(t, err)
    require.NotZero(t, user.ID)
}
```

Example handler test:

```go
func TestUserCreatedHandler(t *testing.T) {
    db := support.NewDatabase(t)

    handler := NewUserCreatedHandler(
        NewUserRepository(db),
    )

    err := handler.Handle(
        t.Context(),
        UserCreated{
            UserID: 1,
        },
    )

    require.NoError(t, err)
}
```

Do not start an HTTP server or a complete ForGo application when the component can be tested directly.

Use the smallest meaningful boundary.

---

# Unit tests

Unit tests do not belong under `/test`.

Keep them next to the production code they test:

```text
internal/
└── domain/
    └── user/
        ├── user.go
        └── user_test.go
```

Unit tests should normally avoid real external infrastructure.

Dependencies may be replaced with fakes, stubs, or other lightweight test implementations where useful.

---

# Choosing the correct test type

Use the following rule of thumb:

```text
Pure function or isolated business logic
    → unit test

Repository/service/handler + real infrastructure
    → integration test

One HTTP endpoint + response + side effects
    → HTTP feature test

Multiple HTTP calls representing a business workflow
    → HTTP E2E test
```

Do not promote a test to a larger scope without a reason.

For example, a repository does not need to be tested through an HTTP endpoint simply because the repository is eventually used by one.

Likewise, a single endpoint does not need to become an E2E flow when its behavior can be completely verified by one request.

---

# Test infrastructure

Real infrastructure should preferably be disposable and reproducible.

For example:

```go
func TestSomething(t *testing.T) {
    db := support.NewDatabase(t)

    // test...
}
```

`support.NewDatabase` may internally use Testcontainers or another test infrastructure mechanism.

Application tests should not depend on developers manually configuring long-lived databases before running them unless there is a deliberate project-specific reason to do so.

Test infrastructure helpers should hide container lifecycle details from individual tests.

Tests should focus on behavior:

```go
db := support.NewDatabase(t)
```

rather than infrastructure mechanics:

```go
container, err := testcontainers.Run(...)
...
```

---

# Test readability

Tests should be readable as behavior without requiring a separate DSL.

Prefer normal Go:

```go
func TestCreateUser(t *testing.T) {
    env := support.NewHTTP(t)

    env.Fixture(t, "users/empty.sql")

    resp := env.Post(t, "/v1/users", CreateUserRequest{
        Email: "user@example.com",
    })

    resp.Status(t, http.StatusOK)

    env.DBHas(t, "users", map[string]any{
        "email": "user@example.com",
    })
}
```

over abstractions such as:

```text
Given(...)
When(...)
Then(...)
```

or external Gherkin feature definitions.

Small domain-oriented helpers are encouraged when they improve readability:

```go
user := register(t, env, "user@example.com")
token := login(t, env, user)
profile := getProfile(t, env, token)
```

The goal is not to minimize the number of lines.

The goal is for a developer or coding agent to understand **what behavior is being verified** without first learning a large testing DSL.

---

# Shared test helpers

Common helpers may be placed in a test support package if needed:

```text
test/
├── support/
│   ├── database.go
│   ├── http.go
│   ├── fixtures.go
│   └── rabbitmq.go
```

Keep these helpers small and focused.

They may handle concerns such as:

- starting and stopping test infrastructure,
- resetting database state,
- loading SQL fixtures,
- constructing the HTTP test server,
- making authenticated requests,
- decoding responses,
- frequently used database assertions.

Do not move application behavior into test helpers.

A helper should make a test easier to express, not hide the behavior being tested.