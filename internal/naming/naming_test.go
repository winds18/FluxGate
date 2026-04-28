package naming

import "testing"

func TestAutoPrefixAddsSequenceInsideBrackets(t *testing.T) {
	got := AutoPrefix("机场A", "", "", 2)
	if got != "[机场A-2] " {
		t.Fatalf("unexpected prefix: %q", got)
	}
}

func TestRawNameFromURIUsesFragment(t *testing.T) {
	got := RawNameFromURI("vless://uuid@example.com:443?security=tls#香港%2001")
	if got != "香港 01" {
		t.Fatalf("unexpected raw name: %q", got)
	}
}

func TestDisplayName(t *testing.T) {
	got := DisplayName("[机场A] ", "香港 01")
	if got != "[机场A] 香港 01" {
		t.Fatalf("unexpected display name: %q", got)
	}
}
