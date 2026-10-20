// Command worker handles messages and relays the outbox.
package main

import (
	"os"

	"example/internal/app"
)

func main() {
	if app.Worker().Run() != nil {
		os.Exit(1) // Run logged the failure
	}
}
