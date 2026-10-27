package features

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// env serves the orders API over gRPC on loopback, as cmd/grpc does, for every test of the package.
var env *forgotest.Env

func TestMain(m *testing.M) {
	stop := support.StartDatabase()
	defer stop()
	env = support.GRPC()
	defer env.Close()
	m.Run()
}
