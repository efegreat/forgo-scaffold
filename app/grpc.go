package app

import (
	"github.com/efegreat/forgo"

	"example/config"
)

// GRPC is the app of the orders API over gRPC, which cmd/grpc runs. It runs the services of HTTP's, and like
// it needs no broker.
func GRPC() *forgo.App {
	return forgo.NewApp(options("example-grpc",
		forgo.WithServers(config.GRPC, config.Ops(":8084", config.Drain)),
	)...)
}
