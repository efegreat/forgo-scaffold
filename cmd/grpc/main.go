// Command grpc serves the orders API over gRPC.
package main

import (
	"os"

	"example/app"
)

func main() {
	if app.GRPC().Run() != nil {
		os.Exit(1) // Run logged the failure
	}
}
