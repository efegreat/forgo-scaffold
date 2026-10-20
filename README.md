# forgo scaffold

A starting point for a backend on [forgo](https://github.com/efegreat/forgo) and [forgo-adapter](https://github.com/efegreat/forgo-adapter): an HTTP API, a worker and scheduled jobs, on PostgreSQL and RabbitMQ, with tests that run the same application `main` runs.

The example domain is small on purpose. Placing an order stores it and saves an `order.placed` message to the outbox in the same transaction. The worker relays the outbox to RabbitMQ, records a shipment for each message, and a job dispatches the pending shipments. Rename or delete `internal/order` and `internal/shipping` to start your own.

```text
.
├── cmd/
│   ├── http/          app.HTTP().Run(): the orders API
│   ├── worker/        app.Worker().Run(): the consumer and the outbox relay
│   ├── cron/          app.Cron().Run(): the scheduled jobs
│   └── cli/           app.CLI().Run(): commands, such as migrate
├── config/            the resources (functions) and the servers (vars)
├── internal/
│   ├── app/           the app of each deployable, shared by main and the tests
│   ├── order/         places orders, and saves their messages to the outbox
│   ├── shipping/      handles order.placed, and the dispatch job
│   └── postgres/      the repositories and the outbox writer
├── migrations/        postgres/ and mysql/, embedded by migrations.go
└── test/              feature, E2E, worker, cron, contract and integration tests (test/README.md)
```

## Run

The deployables read their configuration from the environment:

| Variable | Default |
| --- | --- |
| `DATABASE_URL` | `postgres://example:example@localhost:5432/example?sslmode=disable` |
| `RABBITMQ_HOST`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD` | `localhost`, `guest`, `guest` |
| `HTTP_ADDR` | `:8080` |

```console
$ go run ./cmd/cli migrate -a=up
$ go run ./cmd/http
$ curl -X POST localhost:8080/v1/orders -d '{"sku":"a-1","quantity":2}' -H 'Content-Type: application/json'
$ go run ./cmd/worker
```

`go run ./cmd/cli migrate -a=new -name=create_x` writes a new migration to `migrations/postgres`.

## Test

```console
$ go test ./...
```

The tests need Docker: each test package starts a PostgreSQL container. They run no broker. See [test/README.md](test/README.md).

## Notes

- Until forgo and forgo-adapter are released, `go.mod` reads them from checkouts next to this one (`replace`). Drop the `replace` block once they are.
- The example runs on PostgreSQL. `migrations/mysql` holds forgo's storage tables for MySQL. Switching takes MySQL versions of the example's tables and repositories, and forgo-adapter's MySQL outbox and inbox.
- The broker topology in `config.Broker` is not covered by the tests. Check it against a real broker.
