package store

import (
	"context"
	"path/filepath"
	"testing"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "fluxgate.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSourcePrefixAndNodeDisplayNames(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	sourceA, err := db.CreateSource(ctx, CreateSourceInput{Name: "机场A", Type: "manual"})
	if err != nil {
		t.Fatalf("create source A: %v", err)
	}
	sourceB, err := db.CreateSource(ctx, CreateSourceInput{Name: "机场A", Type: "manual"})
	if err != nil {
		t.Fatalf("create source B: %v", err)
	}
	if sourceA.DisplayPrefix != "[机场A] " {
		t.Fatalf("unexpected source A prefix: %q", sourceA.DisplayPrefix)
	}
	if sourceB.DisplayPrefix != "[机场A-2] " {
		t.Fatalf("unexpected source B prefix: %q", sourceB.DisplayPrefix)
	}

	result, err := db.ImportNodes(ctx, ImportNodesInput{
		SourceID: sourceA.ID,
		Content:  "vless://uuid@example.com:443#香港%2001",
	})
	if err != nil {
		t.Fatalf("import nodes: %v", err)
	}
	if result.Imported != 1 {
		t.Fatalf("expected one imported node, got %+v", result)
	}

	nodes, err := db.ListNodes(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected one node, got %d", len(nodes))
	}
	if nodes[0].DisplayName != "[机场A] 香港 01" {
		t.Fatalf("unexpected display name: %q", nodes[0].DisplayName)
	}

	if _, err := db.UpdateNodeDisplayName(ctx, nodes[0].ID, "手动香港"); err != nil {
		t.Fatalf("manual rename: %v", err)
	}
	if _, err := db.UpdateSourcePrefix(ctx, sourceA.ID, "[新前缀]"); err != nil {
		t.Fatalf("update source prefix: %v", err)
	}
	node, err := db.GetNode(ctx, nodes[0].ID)
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if node.DisplayName != "手动香港" {
		t.Fatalf("manual name should not be overwritten: %q", node.DisplayName)
	}

	node, err = db.ResetNodeDisplayName(ctx, nodes[0].ID)
	if err != nil {
		t.Fatalf("reset node name: %v", err)
	}
	if node.DisplayName != "[新前缀] 香港 01" {
		t.Fatalf("unexpected reset display name: %q", node.DisplayName)
	}
}

func TestTokenCreation(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	team, err := db.CreateTeam(ctx, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	user, err := db.CreateUser(ctx, CreateUserInput{TeamID: &team.ID, Name: "Alice"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	result, err := db.CreateToken(ctx, "secret", "https://flux.example", CreateTokenInput{
		UserID:     user.ID,
		Name:       "Alice 30d",
		ExpireDays: 30,
		QuotaBytes: 1024,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if result.PlainToken == "" || result.Token.TokenPrefix == "" || result.Account.UUID == "" {
		t.Fatalf("incomplete token result: %+v", result)
	}
	if result.Subscription == "" {
		t.Fatalf("subscription URL missing")
	}
}
