package config

import (
	"testing"
	"time"
)

func TestLoadSubStoreExtractionConfig(t *testing.T) {
	t.Setenv("SUB_STORE_EXTRACT_URL_TEMPLATE", "http://sub-store:3001/extract?url={url}")
	t.Setenv("SUB_STORE_TIMEOUT_SECONDS", "7")

	cfg := Load()
	if cfg.SubStoreExtractURLTemplate != "http://sub-store:3001/extract?url={url}" {
		t.Fatalf("unexpected sub-store template: %q", cfg.SubStoreExtractURLTemplate)
	}
	if cfg.SubStoreTimeout != 7*time.Second {
		t.Fatalf("unexpected sub-store timeout: %s", cfg.SubStoreTimeout)
	}
}
