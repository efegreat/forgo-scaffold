package contract

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// api and worker run the two deployables that a message connects, over one database, as two processes
// do in production.
var api, worker *forgotest.Env

func TestMain(m *testing.M) {
	stop := support.StartDatabase()
	defer stop()
	api = support.HTTP()
	defer api.Close()
	worker = support.Worker()
	defer worker.Close()
	m.Run()
}
