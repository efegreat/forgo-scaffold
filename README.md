# forgo scaffold

A starting point for a backend on [forgo](https://github.com/efegreat/forgo) and [forgo-adapter](https://github.com/efegreat/forgo-adapter): an HTTP API, a worker and scheduled jobs, on PostgreSQL, RabbitMQ and Redis, with metrics, live configs, and tests that run the same application `main` runs.

The example domain is small on purpose:

1. Placing an order stores it and saves an `order.placed` message to the outbox, in the same transaction. A live config limits the quantity of an order, and a rate limit the orders of a client.
2. The worker relays the outbox to RabbitMQ, and records a shipment for each `order.placed`.
3. A job dispatches the pending shipments, under a lock that keeps two replicas from dispatching at once, and saves a `shipment.dispatched` message for each.
4. The worker relays it to the order domain, which marks the order shipped.

Replace the `order` and `shipping` domains, and what they declare in the other layers, to start your own.

The layout is an example, not something forgo requires. forgo works with any layout, and its [Project layout](https://github.com/efegreat/forgo/blob/main/REFERENCE.md#project-layout) lists the few rules that come from forgo itself.

```text
.
├── cmd/
│   ├── http/            app.HTTP().Run(): the orders API
│   ├── worker/          app.Worker().Run(): the consumer and the outbox relay
│   ├── cron/            app.Cron().Run(): the scheduled jobs
│   └── cli/             app.CLI().Run(): migrate, openapi and retry-failed
├── app/                 the app of each deployable, one file each, shared by main and the tests
├── config/              the resources (functions) and the servers (vars), one file each
├── internal/
│   ├── domain/          the business logic, one package per domain
│   │   ├── order/       service.go, request.go, response.go, config.go: places, reads and ships orders
│   │   ├── shipping/    service.go, job.go: handles order.placed, and the dispatch job
│   │   └── messaging/   retry.go: the retry-failed command
│   ├── entity/          the business objects the domains share, and their rules: Order, Shipment
│   ├── errs/            the business errors of every domain, as operation.Error values
│   ├── event/           the messages the domains exchange: names, destinations, payloads
│   ├── port/            the interfaces the domains need: repository.go, outbox.go, inbox.go
│   └── adapter/         the implementations of the ports
│       └── postgres/    orders.go, shipments.go, failed_messages.go
├── api/                 the OpenAPI document of the API, which the openapi command writes
├── migrations/          postgres/ and mysql/, embedded by migrations.go
├── layers_test.go       the import rules of internal
├── test/                feature, E2E, worker, cron, contract, CLI and integration tests (test/README.md)
├── Dockerfile           one image of every deployable
└── compose.yaml         the whole system on one machine
```

## Layers

The layers are this scaffold's choice, a hexagonal one, not forgo's. What forgo itself asks for holds in any layout: `app` builds each deployable's app in a function, `config` reads the resources in functions, and the providers come after what they depend on. Change the rest as your project needs. `layers_test.go` checks this layout's rules, so change it with them, or delete it.

`internal` is the application: the domains in the middle, the adapters around them. `cmd`, `app` and `config` compose it into deployables, and `internal` never imports them. Dependencies point inward only:

| Layer | Holds | May import |
| --- | --- | --- |
| `entity` | The business objects, in the business's shape: no tags, no database types. Their methods keep the rules that need nothing else, such as `Order.Ship`. | `errs` |
| `errs` | The business errors, whose codes are the error codes of the API. | |
| `event` | The messages: name, destination, payload, and how to build the message. | |
| `port` | The interfaces the domains need, in terms of `entity` and `event`, one file per kind. | `entity`, `event` |
| `domain/x` | The services, jobs, commands and message handlers of a domain, with their requests, responses and live configs. | `entity`, `errs`, `event`, `port` |
| `adapter/x` | The implementations of the ports, one file per port. | `entity`, `errs`, `event`, `port` |

- Domains never import each other. They share entities, errors and events, and reach other domains through ports and messages, so they never form an import cycle: order and shipping exchange messages both ways.
- Only `adapter` imports infrastructure, such as `database/sql`, forgo-adapter and the clients of other services. `app` provides forgo-adapter's own storages, such as its outbox, which implements `port.Outbox`.
- Adapter constructors return their port, such as `postgres.NewOrders(db) port.OrderRepository`. The compiler checks the implementation, the bootstrap binds it under the port, and a test can put a fake in its place with `forgotest.Replace`.
- `config` is where requests come in: its routes, message handlers, jobs and commands call the domains.

`layers_test.go` fails on an import across these rules, on an import of `app`, `config` or `cmd` from `internal`, and on a directory of `internal` that is no layer.

## Operations

Each deployable serves an **ops server** next to its work, on a port the cluster must not expose: `:8081` for http, `:8082` for worker and `:8083` for cron, or `OPS_ADDR`.

| Route | |
| --- | --- |
| `GET /health` | The liveness probe. It keeps passing while the deployable drains. |
| `GET /metrics` | The metrics of every operation, message, outbox batch and lock, for Prometheus. |
| `GET /live-configs`, `PUT /live-configs/{name}` | The live configs, such as `order_limits`. A change applies to the next order. |

Anyone who reaches the ops port reads every metric and changes every live config. Before exposing it, protect its routes with middlewares, such as `http.JWTAuth` and a role check.

```console
$ curl localhost:8081/live-configs
$ curl -X PUT localhost:8081/live-configs/order_limits -H 'If-Match: 0' -d '{"max_quantity": 5}'
```

The API answers its readiness probe on `GET /ready`, which fails as soon as it starts terminating. It then keeps serving for 10 seconds, so that load balancers stop routing to it before it stops: rolling deployments fail no request. Point the readiness probe at `/ready`, the liveness probe at the ops server's `/health`, and give the pod a termination grace period of 40 seconds (forgo's [zero-downtime deployments](https://github.com/efegreat/forgo/blob/main/REFERENCE.md#zero-downtime-deployments-on-kubernetes)).

The logs mask the values of `config.SensitiveFields` in requests, responses, messages and the logged SQL statements, which are the slow and failed ones.

## Run

The whole system, built from the Dockerfile, with PostgreSQL, RabbitMQ, Redis and Prometheus:

```console
$ docker compose up --build
$ curl -X POST localhost:8080/v1/orders -d '{"sku":"a-1","quantity":2}' -H 'Content-Type: application/json'
```

RabbitMQ's management UI is on `localhost:15672` (guest, guest), and Prometheus on `localhost:9090`.

Or run a deployable on its own, with `go run`. The deployables read their configuration from the environment:

| Variable | Default |
| --- | --- |
| `DATABASE_URL` | `postgres://example:example@localhost:5432/example?sslmode=disable` |
| `RABBITMQ_HOST`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD` | `localhost`, `guest`, `guest` |
| `REDIS_ADDR` | `localhost:6379` |
| `HTTP_ADDR` | `:8080` |
| `OPS_ADDR` | `:8081`, `:8082` or `:8083` |

```console
$ docker compose up -d postgres rabbitmq redis
$ go run ./cmd/cli migrate -a=up
$ go run ./cmd/http
```

The CLI's commands:

| Command | |
| --- | --- |
| `migrate -a=up` | Applies the migrations of `migrations/postgres`. `-a=new -name=create_x` writes a new one there. |
| `openapi -o=api/openapi.json` | Writes the OpenAPI document of the API, `api/openapi-v1.json`. A test fails until it is rewritten after the API changes. |
| `retry-failed -subscriber=order [-message=id]` | Sets the failed messages of a subscriber back to waiting, once the cause of their failure is fixed. Then replay them from its dead-letter queue, such as `order.dlq`, with RabbitMQ's management UI or a shovel. |

## Test

```console
$ go test ./...
```

The tests need Docker: each test package starts its PostgreSQL container, and Redis where it runs the jobs. They run no broker. See [test/README.md](test/README.md). CI runs them, with vet, staticcheck and a build of the image (`.github/workflows/ci.yml`).

## Notes

- Until forgo and forgo-adapter are released, `go.mod` reads them from checkouts next to this one (`replace`), and so do the Dockerfile and CI. Drop the `replace` block, the Dockerfile's `COPY --from` lines and CI's extra checkouts once they are.
- The example runs on PostgreSQL. `migrations/mysql` holds forgo's storage tables for MySQL. Switching takes MySQL versions of the example's tables, an `adapter/mysql` in place of `adapter/postgres`, and forgo-adapter's MySQL outbox and inbox.
- The broker topology in `config.Broker` is not covered by the tests. Check it against a real broker, such as the one of `compose.yaml`.
- The rate limit of `POST /v1/orders` counts per instance. With several instances, count in Redis with `http.RateLimitWith` and forgo-adapter's `ratelimit.NewRedis`.

## License

forgo-scaffold is released under the [MIT License](LICENSE). Replace it with your own license when you start a project from it.
