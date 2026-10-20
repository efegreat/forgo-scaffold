// Command cli runs one command, such as migrate -a=up.
package main

import (
	"os"

	"example/internal/app"
)

func main() {
	if app.CLI().Run() != nil {
		os.Exit(1) // Run logged the failure
	}
}
