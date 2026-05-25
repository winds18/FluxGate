package upstreamsync

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/winds18/FluxGate/internal/store"
	"github.com/winds18/FluxGate/internal/substore"
)

func TestSyncDueRefreshesSubscriptionSources(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "fluxgate.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	source, err := db.CreateSource(ctx, store.CreateSourceInput{
		Name:                   "订阅源A",
		Type:                   "subscription",
		RawContent:             "vless://uuid@example.com:443#香港%2001",
		RefreshIntervalMinutes: 5,
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	summary, err := Refresher{Store: db}.SyncDue(ctx, time.Date(2026, 4, 29, 3, 16, 0, 0, time.UTC), 10, nil)
	if err != nil {
		t.Fatalf("sync due: %v", err)
	}
	if summary.Checked != 1 || summary.Succeeded != 1 || summary.Imported != 1 || summary.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}

	nodes, err := db.ListNodes(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].SourceID != source.ID || nodes[0].DisplayName != "[订阅源A] 香港 01" {
		t.Fatalf("unexpected imported nodes: %+v", nodes)
	}
	updated, err := db.GetSource(ctx, source.ID)
	if err != nil {
		t.Fatalf("get source: %v", err)
	}
	if updated.LastSyncAt == nil || updated.LastError != "" {
		t.Fatalf("expected successful sync metadata: %+v", updated)
	}
}

func TestRefreshSourceUsesSubStoreExtraction(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "fluxgate.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	sourceURL := "https://airport.example/sub?token=secret-value"
	source, err := db.CreateSource(ctx, store.CreateSourceInput{
		Name: "sub-store订阅",
		Type: "subscription",
		URL:  sourceURL,
	})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}

	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if got := r.URL.Query().Get("url"); got != sourceURL {
			t.Fatalf("unexpected forwarded subscription url: %q", got)
		}
		_, _ = w.Write([]byte("vless://1678dd69-bdf8-4468-933e-9e42821fec93@example.com:443#台湾01"))
	}))
	t.Cleanup(server.Close)

	result, err := Refresher{
		Store:    db,
		SubStore: substore.Extractor{URLTemplate: server.URL + "/extract?url={url}"},
	}.RefreshSource(ctx, source)
	if err != nil {
		t.Fatalf("refresh source: %v", err)
	}
	if hits != 1 || result.Imported != 1 {
		t.Fatalf("unexpected sub-store refresh result hits=%d result=%+v", hits, result)
	}
	nodes, err := db.ListNodes(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 || nodes[0].DisplayName != "[sub-store订阅] 台湾01" {
		t.Fatalf("unexpected imported nodes: %+v", nodes)
	}
	updated, err := db.GetSource(ctx, source.ID)
	if err != nil {
		t.Fatalf("get source: %v", err)
	}
	if updated.RawContent != "vless://1678dd69-bdf8-4468-933e-9e42821fec93@example.com:443#台湾01" {
		t.Fatalf("expected extracted raw content, got %q", updated.RawContent)
	}
}
