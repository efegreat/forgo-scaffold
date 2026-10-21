// Command cron runs the scheduled jobs.
package main

import (
	"os"

	"example/app"
)

func main() {
	if app.Cron().Run() != nil {
		os.Exit(1) // Run logged the failure
	}
}
