package upstream

import (
	"encoding/base64"
	"testing"
)

func TestNormalizeContentPlainURIList(t *testing.T) {
	got, err := NormalizeContent(" # comment\nvless://uuid@example.com:443#HK\n\n")
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != "vless://uuid@example.com:443#HK" {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentBase64URIList(t *testing.T) {
	raw := "vless://uuid@example.com:443#HK\nss://example#SG"
	encoded := base64.StdEncoding.EncodeToString([]byte(raw))
	got, err := NormalizeContent(encoded)
	if err != nil {
		t.Fatalf("NormalizeContent returned error: %v", err)
	}
	if got != raw {
		t.Fatalf("unexpected normalized content: %q", got)
	}
}

func TestNormalizeContentRejectsUnsupportedContent(t *testing.T) {
	if _, err := NormalizeContent("not a subscription"); err == nil {
		t.Fatal("expected unsupported content error")
	}
}
