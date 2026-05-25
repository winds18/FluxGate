package substore

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExtractorExtractsNormalizedNodes(t *testing.T) {
	sourceURL := "https://airport.example/sub?token=secret-value"
	rawNodes := strings.Join([]string{
		"vless://1678dd69-bdf8-4468-933e-9e42821fec93@example.com:443#香港01",
		"trojan://password@example.net:443#日本01",
	}, "\n")
	hits := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		hits++
		if got := r.URL.Query().Get("url"); got != sourceURL {
			t.Fatalf("unexpected forwarded subscription url: %q", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(base64.StdEncoding.EncodeToString([]byte(rawNodes)))),
			Header:     make(http.Header),
		}, nil
	})}

	extracted, err := Extractor{
		URLTemplate: "http://sub-store.local/extract?url={url}",
		Client:      client,
	}.Extract(context.Background(), sourceURL)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if hits != 1 {
		t.Fatalf("expected one sub-store request, got %d", hits)
	}
	if extracted != rawNodes {
		t.Fatalf("unexpected extracted nodes:\n%s", extracted)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestExtractorRequiresURLPlaceholder(t *testing.T) {
	_, err := Extractor{URLTemplate: "https://sub-store.example/extract"}.Extract(context.Background(), "https://airport.example/sub")
	if err == nil || !strings.Contains(err.Error(), "{url}") {
		t.Fatalf("expected placeholder error, got %v", err)
	}
}
