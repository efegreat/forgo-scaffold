package contract

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// api, worker and jobs run the deployables that the messages connect, over one database, as separate
// processes do in production.
var api, worker, jobs *forgotest.Env

func TestMain(m *testing.M) {
	stopDB := support.StartDatabase()
	defer stopDB()
	stopRedis := support.StartRedis()
	defer stopRedis()
	api = support.HTTP()
	defer api.Close()
	worker = support.Worker()
	defer worker.Close()
	jobs = support.Cron()
	defer jobs.Close()
	m.Run()
}
