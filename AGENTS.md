# AGENTS.md

Guide for coding agents and contributors working on this application. [README.md](README.md) describes the layout and how to run it, [test/README.md](test/README.md) the tests. forgo's [REFERENCE.md](https://github.com/efegreat/forgo/blob/main/REFERENCE.md) and forgo-adapter's [REFERENCE.md](https://github.com/efegreat/forgo-adapter/blob/main/REFERENCE.md) document the framework and the adapters.

## Layout rules

- `internal` is the application, in layers: `entity`, `errs`, `event`, `port`, `domain/<name>` and `adapter/<name>`. `cmd`, `app` and `config` compose it into deployables. README.md's Layers table says what each layer holds and may import, and `layers_test.go` enforces it: never weaken the test to let an import through, move the code instead.
- Domains never import each other. What two domains share goes in `entity`, `errs` or `event`, and what one needs from outside goes behind a port in `port`.
- Only `adapter` imports infrastructure: `database/sql`, forgo-adapter, the clients of other services. Adapter constructors return their port, never their type, which stays unexported.
- Business code never imports a transport. Requests and responses are plain structs whose tags (`json`, `uri`, `query`, `header`, `validate`) the servers read.
- A business error is an `operation.Error` in `errs`, with a unique `Code`. Its `Message` is what clients read: never put values or the wrapped error's text in it.

## Recipes

**A new endpoint.**
1. Add its request and response to `domain/<name>/request.go` and `response.go`, and its method to the domain's service, shaped `func(ctx, req) (resp, error)`.
2. Route it in `config/http.go` with `http.HandlerFor((*<name>.Service).Method)`.
3. Rewrite the API document with `go run ./cmd/cli openapi -o=api/openapi.json`.
4. Add a feature test under `test/http/features`.
5. To serve it over gRPC too, add its rpc and messages to the domain's `.proto` file, such as `api/order/v1/order.proto`, with the fields of its request and response, and rewrite the Go code with `go run github.com/bufbuild/buf/cmd/buf@v1.75.0 generate`. Route it in `config/grpc.go` with `fgrpc.HandlerFor((*<name>.Service).Method)`, named like its HTTP route, and add a test under `test/grpc/features`.

**A new message.**
1. Declare its name, destination and payload, and its `Message` method, in `internal/event`.
2. Save it to the outbox with `port.Outbox`, in the transaction of the change it announces.
3. Handle it in the consuming domain, and route it in `config/worker.go` under its subscriber.
4. Declare the exchange, queue, binding and subscriber in `config.Broker`.
5. Add a worker test, and a contract test from the operation that saves it to the handler.

**A new job.** Write it in its domain (`job.go`), shaped as a `cron.IJob` with its `Tolerance`. Declare it in `config/cron.go`, with a `Lock` whose `TTL` is above its longest run. Add a cron test.

**A new command.** Write it in its domain, as a `cli.CommandFactory` for `cli.As`, and declare it in `app/cli.go`. Add a test under `test/cli`.

**A new live config.** Declare it in its domain's `config.go`, with `Name`, `Default` and `validate` tags, and add it to `config.LiveConfigs` so that admins can change it.

**A new table or index.** Write a migration with `go run ./cmd/cli migrate -a=new -name=<name>`, never edit an applied one, and add fixtures under `test/fixtures` for the scenarios that need it.

**A new provider.** Add it to the shared `providers` of `app/app.go`, after what it depends on. A provider that needs a resource only some deployables have goes to those apps instead, with `forgo.WithProviders`.

## Tests

- Tests run the app that `main` runs, through forgo's `forgotest`, against real PostgreSQL and Redis containers. Never add a second wiring for tests, and prefer configuration to `forgotest.Replace`.
- Each test package builds its env once, in `TestMain`. Tests start with `support.Fixture` and never call `t.Parallel`.
- Test behavior that the code owns: a rule, a status, a message saved, a lock honored. A setting alone, such as a schedule, is configuration.

## Workflow

```sh
go build ./...
go vet ./...
go test -race ./...   # needs Docker
gofmt -l .            # must print nothing
go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
```

CI (`.github/workflows/ci.yml`) runs these, `go mod tidy -diff`, a check that the code generated from `api`'s `.proto` files is current, and a build of the image.
