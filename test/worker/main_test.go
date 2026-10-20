package worker

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// env runs the worker, as cmd/worker does but without RabbitMQ, for every test of the package. It
// consumes nothing: tests hand it messages with env.Deliver.
var env *forgotest.Env

func TestMain(m *testing.M) {
	stop := support.StartDatabase()
	defer stop()
	env = support.Worker()
	defer env.Close()
	m.Run()
}
