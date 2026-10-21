package cli

import (
	"testing"

	"example/test/support"
)

// retry-failed takes the subscriber whose messages to retry, and flags only. The worker's tests retry a
// failed message through it.
func TestRetryFailedFlags(t *testing.T) {
	for _, args := range [][]string{
		{"retry-failed"},
		{"retry-failed", "-message=m1"},
		{"retry-failed", "-subscriber=order", "m1"},
	} {
		if err := support.Command(args...); err == nil {
			t.Errorf("%v ran", args)
		}
	}
	support.Fixture(t, "inbox/failed_messages.sql")
	if err := support.Command("retry-failed", "-subscriber=shipping"); err != nil {
		t.Fatal(err)
	}
	if n := support.Count(t, `SELECT COUNT(*) FROM inbox_events WHERE subscriber_name = 'shipping' AND status = 0`); n != 2 {
		t.Fatalf("%d messages of shipping waiting, want its two failed ones", n)
	}
}
