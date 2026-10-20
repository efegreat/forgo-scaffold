package integration

import (
	"testing"

	"example/test/support"
)

func TestMain(m *testing.M) {
	stop := support.StartDatabase()
	defer stop()
	m.Run()
}
