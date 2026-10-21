package cli

import (
	"testing"

	"example/test/support"
)

func TestMain(m *testing.M) {
	stop := support.StartDatabase() // every command opens the database first
	defer stop()
	m.Run()
}
