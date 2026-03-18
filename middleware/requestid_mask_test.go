package middleware

import "testing"

func TestMaskedServerIP(t *testing.T) {
	t.Parallel()

	if got := maskedServerIP("10.20.30.40"); got != "10.20.*.40" {
		t.Fatalf("unexpected masked ip: %s", got)
	}
	if got := maskedServerIP("127.0.0.1"); got != "127.0.*.1" {
		t.Fatalf("unexpected localhost mask: %s", got)
	}
	if got := maskedServerIP("invalid"); got != "invalid" {
		t.Fatalf("unexpected non-ip mask: %s", got)
	}
}
