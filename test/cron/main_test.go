package cron

import (
	"testing"

	"github.com/efegreat/forgo/forgotest"

	"example/test/support"
)

// env runs the scheduled jobs, as cmd/cron does, for every test of the package. It schedules nothing:
// tests run a job once with env.RunJob.
var env *forgotest.Env

func TestMain(m *testing.M) {
	stopDB := support.StartDatabase()
	defer stopDB()
	stopRedis := support.StartRedis()
	defer stopRedis()
	env = support.Cron()
	defer env.Close()
	m.Run()
}
