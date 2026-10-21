# forgo scaffold

A starting point for a backend on [forgo](https://github.com/efegreat/forgo) and [forgo-adapter](https://github.com/efegreat/forgo-adapter): an HTTP API, a worker and scheduled jobs, on PostgreSQL and RabbitMQ, with tests that run the same application `main` runs.

The example domain is small on purpose. Placing an order stores it and saves an `order.placed` message to the outbox in the same transaction. The worker relays the outbox to RabbitMQ, records a shipment for each message, and a job dispatches the pending shipments. Replace the `order` and `shipping` domains, and what they declare in the other layers, to start your own.

```text
.
├── cmd/
│   ├── http/          app.HTTP().Run(): the orders API
│   ├── worker/        app.Worker().Run(): the consumer and the outbox relay
│   ├── cron/          app.Cron().Run(): the scheduled jobs
│   └── cli/           app.CLI().Run(): commands, such as migrate
├── config/            the resources (functions) and the servers (vars), one file each
├── internal/
│   ├── app/           the app of each deployable, one file each, shared by main and the tests
│   ├── domain/        the business logic, one package per domain
│   │   ├── order/     service.go, request.go, response.go: places and reads orders
│   │   └── shipping/  service.go, job.go: handles order.placed, and the dispatch job
│   ├── entity/        the business objects the domains share: Order, Shipment
│   ├── errs/          the business errors of every domain, as operation.Error values
│   ├── event/         the messages the domains exchange: names, destinations, payloads
│   ├── port/          the interfaces the domains need: repository.go, outbox.go
│   └── adapter/       the implementations of the ports
│       └── postgres/  orders.go, shipments.go
├── migrations/        postgres/ and mysql/, embedded by migrations.go
├── layers_test.go     the import rules of internal
└── test/              feature, E2E, worker, cron, contract and integration tests (test/README.md)
```

## Layers

The domains are in the middle, the adapters around them, and `app` wires them together. Dependencies point inward only:

| Layer | Holds | May import |
| --- | --- | --- |
| `entity` | The business objects, in the business's shape: no tags, no database types. | `errs` |
| `errs` | The business errors, whose codes are the error codes of the API. | |
| `event` | The messages: name, destination, payload, and how to build the message. | |
| `port` | The interfaces the domains need, in terms of `entity` and `event`, one file per kind. | `entity`, `event` |
| `domain/x` | The services, jobs and message handlers of a domain, with their requests and responses. | `entity`, `errs`, `event`, `port` |
| `adapter/x` | The implementations of the ports, one file per port. | `entity`, `errs`, `event`, `port` |
| `app` | The app of each deployable: its providers, resources and servers. | all of them |

- Domains never import each other. They share entities, errors and events, and reach other domains through ports and messages, so they never form an import cycle.
- Only `adapter` and `app` import infrastructure, such as `database/sql`, forgo-adapter and the clients of other services.
- A port can be implemented outside `adapter`: forgo-adapter's outbox implements `port.Outbox`, so the app provides it directly.
- Adapter constructors return their port, such as `postgres.NewOrders(db) port.OrderRepository`. The compiler checks the implementation, the bootstrap binds it under the port, and a test can put a fake in its place with `forgotest.Replace`.
- `config` is where requests come in: its routes, message handlers and jobs call the domains.

`layers_test.go` fails on an import across these rules, and on a directory of `internal` that is no layer.

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
- The example runs on PostgreSQL. `migrations/mysql` holds forgo's storage tables for MySQL. Switching takes MySQL versions of the example's tables, an `adapter/mysql` in place of `adapter/postgres`, and forgo-adapter's MySQL outbox and inbox.
- The broker topology in `config.Broker` is not covered by the tests. Check it against a real broker.
