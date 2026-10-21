// Package config declares the resources and servers of the application, one file each: resources.go, and
// the servers of each deployable.
//
// The resources are functions, not vars: Go evaluates package-level vars before TestMain, so the address
// of a test's database container, which TestMain puts in DATABASE_URL, could never reach them. The
// servers can stay vars, since tests serve them on ports of their own.
package config

import (
	"github.com/efegreat/forgo/logger"
	"github.com/rs/zerolog"
)

// Logger writes JSON lines to stdout.
var Logger = logger.Config{Level: zerolog.InfoLevel}
