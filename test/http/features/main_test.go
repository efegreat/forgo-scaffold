package features

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// env runs the orders API, as cmd/http does, for every test of the package.
var env *forgotest.Env

func TestMain(m *testing.M) {
	stop := support.StartDatabase()
	defer stop()
	env = support.HTTP()
	defer env.Close()
	m.Run()
}
