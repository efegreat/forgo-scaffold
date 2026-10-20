# Testing

This directory contains tests that exercise the application beyond isolated unit tests.

The general rule is:

- Unit tests stay next to the code they test.
- Tests under `/test` use real infrastructure where it can run in a test, such as PostgreSQL in a disposable container started through Testcontainers.
- Tests that go through a server run the application that `main` runs: `internal/app` builds it, and forgo's [`forgotest`](https://github.com/efegreat/forgo/blob/main/REFERENCE.md#testing) runs it. There is no second wiring for tests.
- Tests remain ordinary Go tests. Avoid introducing Gherkin-style abstractions unless there is a strong reason to do so.

```text
test/
├── fixtures/          SQL fixtures by scenario, embedded by fixtures.go
├── support/           the database container, the envs, fixtures, an HTTP client
├── http/
│   ├── features/      one endpoint, through the router in process
│   └── e2e/           a flow of several requests, over the network
├── worker/            message handlers, through env.Deliver
├── cron/              jobs, through env.RunJob
├── contract/          a message, from the endpoint that saves it to the handler that handles it
└── integration/       components against real infrastructure, without an app
```

Run them all with `go test ./...`. They need Docker, and nothing else.

---

# One env per package

A test package starts its database and builds its env once, in `TestMain`, and its tests share them:

```go
// test/http/features/main_test.go
var env *forgotest.Env

func TestMain(m *testing.M) {
    stop := support.StartDatabase() // a container, migrated with the app's own migrate command
    defer stop()
    env = support.HTTP()            // app.HTTP(), as cmd/http runs it
    defer env.Close()
    m.Run()
}
```

- `support.StartDatabase` starts the PostgreSQL container of the package and points `DATABASE_URL` at it, before any env reads it. That is why `config.DB` is a function, not a var: Go evaluates package-level vars before `TestMain`.
- `support.HTTP`, `support.Worker` and `support.Cron` build the env of each deployable of `internal/app`.
- The tests of a package share the env and the database, so they do not call `t.Parallel()`. Each test resets the database first, with `support.Fixture`. Packages still run in parallel, each with its own container.

---

## `fixtures/`

Contains reusable test data used to establish an application's initial state.

SQL fixtures are preferred for relational databases because they are explicit, easy to inspect, and support database-specific behavior without introducing another data representation.

```text
test/
└── fixtures/
    ├── fixtures.go
    ├── orders/
    │   └── placed_order.sql
    └── shipments/
        └── pending_shipments.sql
```

A fixture should describe a useful test state rather than mirror individual database tables.

Prefer:

```text
placed_order.sql
pending_shipments.sql
expired_session.sql
```

over:

```text
orders.sql
shipments.sql
sessions.sql
```

when the fixture exists specifically to prepare a business scenario.

`support.Fixture(t, names...)` empties every table but the migrations, the outbox and the inbox included, then runs the named fixtures in order. With no names, it leaves an empty database:

```go
support.Fixture(t)                                    // empty
support.Fixture(t, "shipments/pending_shipments.sql") // two pending shipments, one dispatched
```

Fixtures may be shared between feature, E2E, worker, cron, contract and integration tests.

---

# HTTP tests

Tests under `http/` exercise the application through its HTTP boundary.

They use the real HTTP configuration, `config.HTTP`, including:

- routing
- middleware
- request binding
- validation
- handlers
- serialization
- error mapping

Application dependencies used by the request path are backed by real test infrastructure.

HTTP tests are divided into **features** and **E2E** tests based on the scope of behavior being tested. Both run the whole chain of the route. They differ in how a request reaches it:

- `support.Router(t, env)` serves requests through the router of the HTTP server, in process, without the network. A request costs a couple of microseconds.
- `support.Live(t, env)` sends them to the server over loopback, as clients do, through net/http's server: its timeouts and limits, and connection handling.

## `http/features/`

Feature tests verify one HTTP endpoint or one closely related HTTP behavior.

A feature test normally checks:

1. the initial state,
2. one HTTP request,
3. the HTTP response,
4. relevant side effects, such as the rows written and the messages saved to the outbox.

```go
func TestPlaceOrder(t *testing.T) {
    support.Fixture(t)

    resp := support.Router(t, env).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2})

    resp.Status(t, http.StatusOK)
    placed := support.Result[order.PlaceResp](t, resp)
    if n := support.Count(t, `SELECT COUNT(*) FROM orders WHERE id = $1`, placed.ID); n != 1 {
        t.Fatalf("%d orders stored", n)
    }

    msgs, err := env.Outbox(t.Context()) // what the endpoint saved for the worker to publish
    // ...
}
```

A feature test should generally stay focused on a single endpoint.

Good examples:

```text
place_order_test.go
get_order_test.go
login_test.go
cancel_order_test.go
```

Avoid turning one feature test into a complete business journey involving many unrelated endpoints. That belongs under `e2e/`.

---

## `http/e2e/`

E2E tests exercise a business flow through multiple HTTP requests.

They verify that independently exposed HTTP operations work correctly when used together in the way a real client or business process would use them.

```text
place order
    ↓
get order
```

```go
func TestPlaceAndGetOrder(t *testing.T) {
    support.Fixture(t)
    client := support.Live(t, env)

    id := placeOrder(t, client, order.PlaceReq{SKU: "a-1", Quantity: 2})
    got := getOrder(t, client, id)

    if got.Status != order.StatusPlaced {
        t.Fatalf("order %+v", got)
    }
}
```

Helpers such as `placeOrder` and `getOrder` still perform real HTTP requests.

Their purpose is to remove repetitive HTTP plumbing so the test itself reads like the business flow being verified.

E2E tests should not bypass HTTP by directly calling services for intermediate steps.

If the test begins as an HTTP E2E test, the important business interactions in that scenario should continue through the HTTP boundary.

E2E tests do **not** mean testing the complete production deployment.

They run the real forgo HTTP server on a loopback port, with real backing infrastructure.

The purpose is to verify the externally observable HTTP workflow, not Kubernetes, ingress, DNS, or production networking.

---

# Worker and cron tests

The env of the worker and the env of the jobs never consume from a broker or run a schedule. Tests run one operation at a time instead, through the same chain production runs.

## `worker/`

`env.Deliver` hands a message to a subscriber, as the broker would, and returns how the consumer settled it:

```go
got, err := env.Deliver(t.Context(), "shipping", msg)
// got.Settlement: forgotest.Completed, Retried (got.Attempt) or Rejected
```

It runs the handler of the message's name, the binding and the validation of its payload, the inbox and the transaction, and the retry rules. So delivering a message twice shows that its order ships once, and a payload the handler's validation rejects is rejected.

Tests run no broker: `support.Worker` puts forgotest's stand-in in place of RabbitMQ.

## `cron/`

`env.RunJob` runs a job once, now, through its chain:

```go
err := env.RunJob(t.Context(), "dispatch_shipments")
```

Whether the schedule is the one you meant is configuration, not behavior, so tests never run the scheduler.

---

# `contract/`

A message connects two deployables: the endpoint that saves it to the outbox, and the handler that handles it. The contract test runs both, as two envs over one database, just as two processes run in production:

```go
resp := support.Router(t, api).Post(t, "/v1/orders", order.PlaceReq{SKU: "a-1", Quantity: 2})
msgs, err := api.Outbox(t.Context())                         // what the relay would publish
got, err := worker.Deliver(t.Context(), "shipping", msgs[0]) // as the broker would deliver it
```

It catches a message name that no handler is registered under, and a payload that the handler's validation rejects.

What it cannot catch is the broker between them. RabbitMQ has to route the exchange `orders` to the queue `shipping`, as `config.Broker` declares. A wrong binding loses messages in production while every test passes. **The broker topology is the application's responsibility.** Verify it where it is declared, against a real broker, in staging or a smoke check.

---

# `integration/`

Integration tests verify a particular component together with real infrastructure.

Unlike HTTP tests, they do not need to exercise a server or build an app.

Typical targets include:

- repositories against a real database,
- services using a real database or Redis instance,
- infrastructure adapters,
- commands that depend on real infrastructure.

```go
func TestOrders(t *testing.T) {
    support.Fixture(t)
    orders := postgres.NewOrders(support.DB())

    err := orders.Insert(t.Context(), order.Order{ID: "…", SKU: "a-1", Quantity: 2, Status: order.StatusPlaced})
    // ...
}
```

Do not start a server or a complete forgo application when the component can be tested directly.

Use the smallest meaningful boundary.

---

# Unit tests

Unit tests do not belong under `/test`.

Keep them next to the production code they test:

```text
internal/
└── order/
    ├── order.go
    └── order_test.go
```

Unit tests should normally avoid real external infrastructure.

Dependencies may be replaced with fakes, stubs, or other lightweight test implementations where useful.

---

# Choosing the correct test type

Use the following rule of thumb:

```text
Pure function or isolated business logic
    → unit test

Repository/service + real infrastructure
    → integration test

One HTTP endpoint + response + side effects
    → HTTP feature test

Multiple HTTP calls representing a business workflow
    → HTTP E2E test

One message handler + side effects
    → worker test

One job run + side effects
    → cron test

A message from the endpoint that saves it to the handler that handles it
    → contract test
```

Do not promote a test to a larger scope without a reason.

For example, a repository does not need to be tested through an HTTP endpoint simply because the repository is eventually used by one.

Likewise, a single endpoint does not need to become an E2E flow when its behavior can be completely verified by one request.

---

# What tests replace

Tests run as much of the production code as they can, so prefer configuration to replacement:

| Dependency | Example | In tests |
| --- | --- | --- |
| Infrastructure that runs in a container | PostgreSQL | Nothing is replaced: `config.DB` reads the container's `DATABASE_URL`. |
| Broker | RabbitMQ | forgotest's stand-in broker, through `forgotest.ReplaceResource` in `support.Worker`. |
| Third-party services | Payments, email, SMS | Preferably an `httptest.Server` fake, whose URL the client's configuration reads. Otherwise, a fake in place of the provider of the client, through `forgotest.Replace`. |

A fake that the tests of a package share is reset by each test, like the database.

---

# Test infrastructure

Real infrastructure is disposable and reproducible. `support.StartDatabase` starts a PostgreSQL container through Testcontainers, migrates it with the app's own `migrate` command, and stops it when `TestMain` returns.

Application tests do not depend on developers manually configuring long-lived databases before running them.

Test infrastructure helpers hide container lifecycle details from individual tests.

Tests focus on behavior:

```go
stop := support.StartDatabase()
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
func TestDispatchShipments(t *testing.T) {
    support.Fixture(t, "shipments/pending_shipments.sql")

    if err := env.RunJob(t.Context(), "dispatch_shipments"); err != nil {
        t.Fatal(err)
    }

    if n := support.Count(t, `SELECT COUNT(*) FROM shipments WHERE status = 'pending'`); n != 0 {
        t.Fatalf("%d shipments still pending", n)
    }
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
id := placeOrder(t, client, order.PlaceReq{SKU: "a-1", Quantity: 2})
got := getOrder(t, client, id)
```

The goal is not to minimize the number of lines.

The goal is for a developer or coding agent to understand **what behavior is being verified** without first learning a large testing DSL.

---

# Shared test helpers

Common helpers live in `support/`:

```text
test/
└── support/
    ├── support.go     the database container, the envs, Fixture and Count
    └── http.go        Router, Live, and the responses' JSON envelope
```

Keep these helpers small and focused.

They handle concerns such as:

- starting and stopping test infrastructure,
- building the envs of the deployables,
- resetting database state and loading SQL fixtures,
- making requests and decoding responses,
- frequently used database checks.

Do not move application behavior into test helpers.

A helper should make a test easier to express, not hide the behavior being tested.
