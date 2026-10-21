// Command http serves the orders API.
package main

import (
	"os"

	"example/app"
)

func main() {
	if app.HTTP().Run() != nil {
		os.Exit(1) // Run logged the failure
	}
}
