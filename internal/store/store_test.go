package store

import (
	"context"
	"path/filepath"
	"strings"
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

func TestImportNodesMarksMissingSubscriptionNodesInactive(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	source, err := db.CreateSource(ctx, CreateSourceInput{Name: "订阅源A", Type: "subscription"})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	result, err := db.ImportNodes(ctx, ImportNodesInput{
		SourceID: source.ID,
		Content: strings.Join([]string{
			"vless://uuid-a@example.com:443#香港%2001",
			"vless://uuid-b@example.net:443#新加坡%2001",
		}, "\n"),
		MarkMissingInactive: true,
	})
	if err != nil {
		t.Fatalf("initial import: %v", err)
	}
	if result.Imported != 2 || result.Inactivated != 0 {
		t.Fatalf("unexpected initial import result: %+v", result)
	}

	result, err = db.ImportNodes(ctx, ImportNodesInput{
		SourceID:            source.ID,
		Content:             "vless://uuid-b@example.net:443#新加坡%2001",
		MarkMissingInactive: true,
	})
	if err != nil {
		t.Fatalf("refresh import: %v", err)
	}
	if result.Updated != 1 || result.Inactivated != 1 {
		t.Fatalf("unexpected refresh import result: %+v", result)
	}

	nodes, err := db.ListNodes(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	statusByRawName := map[string]string{}
	for _, node := range nodes {
		statusByRawName[node.RawName] = node.Status
	}
	if statusByRawName["香港 01"] != "inactive" {
		t.Fatalf("missing node should be inactive: %+v", statusByRawName)
	}
	if statusByRawName["新加坡 01"] != "active" {
		t.Fatalf("seen node should stay active: %+v", statusByRawName)
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

func TestTokenLifecycleOperations(t *testing.T) {
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
		Name:       "Alice lifecycle",
		ExpireDays: 1,
		QuotaBytes: 1024,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if result.Token.ExpireAt == nil {
		t.Fatal("created token should have expire_at")
	}

	extended, err := db.ExtendToken(ctx, result.Token.ID, 30)
	if err != nil {
		t.Fatalf("extend token: %v", err)
	}
	if extended.ExpireAt == nil || !extended.ExpireAt.After(result.Token.ExpireAt.AddDate(0, 0, 29)) {
		t.Fatalf("token was not extended from existing expiry: before=%v after=%v", result.Token.ExpireAt, extended.ExpireAt)
	}
	if _, err := db.ExtendToken(ctx, result.Token.ID, 0); err == nil {
		t.Fatal("zero-day extension should fail")
	}

	quota, err := db.AddTokenQuota(ctx, result.Token.ID, 2048)
	if err != nil {
		t.Fatalf("add token quota: %v", err)
	}
	if quota.QuotaBytes != 3072 {
		t.Fatalf("unexpected quota after add: %d", quota.QuotaBytes)
	}
	if _, err := db.AddTokenQuota(ctx, result.Token.ID, 0); err == nil {
		t.Fatal("zero quota add should fail")
	}

	revoked, err := db.RevokeToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("revoke token: %v", err)
	}
	if revoked.Status != "revoked" || revoked.RevokedAt == nil {
		t.Fatalf("token should be revoked: %+v", revoked)
	}
	account, err := db.GetGatewayAccountByToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get revoked gateway account: %v", err)
	}
	if account.Status != "revoked" {
		t.Fatalf("gateway account should be revoked: %+v", account)
	}

	restored, err := db.RestoreToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("restore token: %v", err)
	}
	if restored.Status != "active" || restored.RevokedAt != nil {
		t.Fatalf("token should be active after restore: %+v", restored)
	}
	account, err = db.GetGatewayAccountByToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get restored gateway account: %v", err)
	}
	if account.Status != "active" {
		t.Fatalf("gateway account should be active after restore: %+v", account)
	}
}

func TestBootstrapAndAuthenticateAdmin(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	bootstrap, err := db.BootstrapAdmin(ctx, "admin", "admin-password")
	if err != nil {
		t.Fatalf("bootstrap admin: %v", err)
	}
	if !bootstrap.Created || bootstrap.Admin.Username != "admin" {
		t.Fatalf("unexpected bootstrap result: %+v", bootstrap)
	}

	if _, err := db.AuthenticateAdmin(ctx, "admin", "admin-password"); err != nil {
		t.Fatalf("authenticate admin: %v", err)
	}
	if _, err := db.AuthenticateAdmin(ctx, "admin", "wrong-password"); err == nil {
		t.Fatal("wrong password should fail")
	}

	second, err := db.BootstrapAdmin(ctx, "other", "other-password")
	if err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	if !second.Skipped || second.Created {
		t.Fatalf("second bootstrap should skip: %+v", second)
	}
}
