// Package config declares the resources and servers of the application, one file each: resources.go, ops.go,
// and the servers of each deployable.
//
// The resources are functions, not vars: Go evaluates package-level vars before TestMain, so the address
// of a test's database container, which TestMain puts in DATABASE_URL, could never reach them. The
// servers can stay vars, since tests serve them on ports of their own.
package config

import (
	"time"

	"github.com/efegreat/forgo/logger"
	"github.com/rs/zerolog"
)

// Logger writes JSON lines to stdout.
var Logger = logger.Config{Level: zerolog.InfoLevel}

// SensitiveFields are the fields whose values the logs mask as [*]: in the requests and responses of the HTTP
// and gRPC servers, in the messages of the consumer, and in the arguments of the logged SQL statements.
// Credential headers, such as Authorization, are always masked. Add the fields of your own that must not
// reach logs.
var SensitiveFields = []string{"password", "secret", "token", "email", "phone"}

// Drain is how long the APIs, HTTP and GRPC, keep serving once they start terminating, while load balancers
// stop routing to them: longer than the readiness probe takes to fail, and than the proxies take to drop the
// instance.
const Drain = 10 * time.Second
