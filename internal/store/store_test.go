package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestImportNodesAppliesSourceDefaultTags(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	source, err := db.CreateSource(ctx, CreateSourceInput{Name: "Tagged Source", Type: "manual", DefaultTags: `["HK","Premium"]`})
	if err != nil {
		t.Fatalf("create source: %v", err)
	}
	if _, err := db.ImportNodes(ctx, ImportNodesInput{
		SourceID: source.ID,
		Content:  "vless://uuid@example.com:443#香港%2001",
	}); err != nil {
		t.Fatalf("import nodes: %v", err)
	}

	nodes, err := db.ListNodes(ctx)
	if err != nil {
		t.Fatalf("list nodes: %v", err)
	}
	if len(nodes) != 1 || !sameStringSet(nodes[0].Tags, []string{"HK", "Premium"}) {
		t.Fatalf("source default tags should be attached to imported node: %+v", nodes)
	}
}

func TestListDueSubscriptionSources(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)
	now := time.Date(2026, 4, 29, 3, 16, 0, 0, time.UTC)

	neverSynced, err := db.CreateSource(ctx, CreateSourceInput{Name: "never", Type: "subscription", RawContent: "vless://uuid@example.com:443#A", RefreshIntervalMinutes: 15})
	if err != nil {
		t.Fatalf("create never synced source: %v", err)
	}
	oldSync, err := db.CreateSource(ctx, CreateSourceInput{Name: "old", Type: "subscription", RawContent: "vless://uuid@example.com:443#B", RefreshIntervalMinutes: 15})
	if err != nil {
		t.Fatalf("create old source: %v", err)
	}
	recentSync, err := db.CreateSource(ctx, CreateSourceInput{Name: "recent", Type: "subscription", RawContent: "vless://uuid@example.com:443#C", RefreshIntervalMinutes: 15})
	if err != nil {
		t.Fatalf("create recent source: %v", err)
	}
	manual, err := db.CreateSource(ctx, CreateSourceInput{Name: "manual", Type: "manual", RawContent: "vless://uuid@example.com:443#D", RefreshIntervalMinutes: 15})
	if err != nil {
		t.Fatalf("create manual source: %v", err)
	}
	inactive, err := db.CreateSource(ctx, CreateSourceInput{Name: "inactive", Type: "subscription", RawContent: "vless://uuid@example.com:443#E", RefreshIntervalMinutes: 15})
	if err != nil {
		t.Fatalf("create inactive source: %v", err)
	}
	disabledInterval, err := db.CreateSource(ctx, CreateSourceInput{Name: "disabled", Type: "subscription", RawContent: "vless://uuid@example.com:443#F", RefreshIntervalMinutes: 0})
	if err != nil {
		t.Fatalf("create disabled interval source: %v", err)
	}

	setLastSync := func(id int64, value time.Time) {
		t.Helper()
		if _, err := db.db.ExecContext(ctx, "UPDATE upstream_sources SET last_sync_at = ? WHERE id = ?", value.UTC().Format("2006-01-02 15:04:05"), id); err != nil {
			t.Fatalf("set last sync: %v", err)
		}
	}
	setLastSync(oldSync.ID, now.Add(-16*time.Minute))
	setLastSync(recentSync.ID, now.Add(-14*time.Minute))
	if _, err := db.db.ExecContext(ctx, "UPDATE upstream_sources SET status = 'inactive' WHERE id = ?", inactive.ID); err != nil {
		t.Fatalf("set inactive: %v", err)
	}

	due, err := db.ListDueSubscriptionSources(ctx, now, 10)
	if err != nil {
		t.Fatalf("list due sources: %v", err)
	}
	got := map[int64]bool{}
	for _, source := range due {
		got[source.ID] = true
	}
	if !got[neverSynced.ID] || !got[oldSync.ID] {
		t.Fatalf("expected never synced and old source to be due: %+v", due)
	}
	for _, source := range []Source{recentSync, manual, inactive, disabledInterval} {
		if got[source.ID] {
			t.Fatalf("source should not be due: %+v in %+v", source, due)
		}
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

func TestPolicyCreation(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	team, err := db.CreateTeam(ctx, CreateTeamInput{Name: "Policy Team"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	policy, err := db.CreatePolicy(ctx, CreatePolicyInput{
		Name:      "团队默认策略",
		ScopeType: "team",
		ScopeID:   &team.ID,
		MaxNodes:  3,
	})
	if err != nil {
		t.Fatalf("create policy: %v", err)
	}
	if policy.ID == 0 || policy.ScopeID == nil || *policy.ScopeID != team.ID || policy.MaxNodes != 3 {
		t.Fatalf("unexpected policy: %+v", policy)
	}
	if policy.IncludeTags != "[]" || policy.ExcludeTags != "[]" || policy.AllowedVirtualNodes != "[]" {
		t.Fatalf("default policy arrays should be empty JSON arrays: %+v", policy)
	}

	policies, err := db.ListPolicies(ctx)
	if err != nil {
		t.Fatalf("list policies: %v", err)
	}
	if len(policies) != 1 || policies[0].Name != "团队默认策略" {
		t.Fatalf("unexpected policy list: %+v", policies)
	}

	overview, err := db.Overview(ctx, "test")
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if overview.Policies != 1 {
		t.Fatalf("overview should include policies: %+v", overview)
	}

	if _, err := db.CreatePolicy(ctx, CreatePolicyInput{Name: "坏策略", ScopeType: "invalid"}); err == nil {
		t.Fatal("invalid scope_type should fail")
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

func TestRecordTrafficSamplesUpdatesTokenUsageAndRollups(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	team, err := db.CreateTeam(ctx, CreateTeamInput{Name: "Traffic Team"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	user, err := db.CreateUser(ctx, CreateUserInput{TeamID: &team.ID, Name: "Traffic User"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	result, err := db.CreateToken(ctx, "secret", "https://flux.example", CreateTokenInput{
		UserID:     user.ID,
		Name:       "traffic token",
		ExpireDays: 30,
		QuotaBytes: 1024,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	sampledAt := time.Date(2026, 4, 29, 6, 7, 0, 0, time.UTC)
	first, err := db.RecordTrafficSample(ctx, RecordTrafficSampleInput{
		SampledAt:        sampledAt,
		MetricType:       "user",
		MetricName:       result.Account.AuthUser,
		RawValueUpload:   100,
		RawValueDownload: 200,
	})
	if err != nil {
		t.Fatalf("record first sample: %v", err)
	}
	if first.UploadBytesDelta != 0 || first.DownloadBytesDelta != 0 {
		t.Fatalf("first sample should establish baseline: %+v", first)
	}

	second, err := db.RecordTrafficSample(ctx, RecordTrafficSampleInput{
		SampledAt:        sampledAt.Add(5 * time.Minute),
		MetricType:       "user",
		MetricName:       result.Account.AuthUser,
		RawValueUpload:   175,
		RawValueDownload: 260,
	})
	if err != nil {
		t.Fatalf("record second sample: %v", err)
	}
	if second.UploadBytesDelta != 75 || second.DownloadBytesDelta != 60 {
		t.Fatalf("unexpected second delta: %+v", second)
	}

	reset, err := db.RecordTrafficSample(ctx, RecordTrafficSampleInput{
		SampledAt:        sampledAt.Add(10 * time.Minute),
		MetricType:       "user",
		MetricName:       result.Account.AuthUser,
		RawValueUpload:   10,
		RawValueDownload: 20,
	})
	if err != nil {
		t.Fatalf("record reset sample: %v", err)
	}
	if reset.UploadBytesDelta != 10 || reset.DownloadBytesDelta != 20 {
		t.Fatalf("counter reset should count current raw values: %+v", reset)
	}

	token, err := db.GetToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if token.UsedUploadBytes != 85 || token.UsedDownloadBytes != 80 {
		t.Fatalf("unexpected token usage: %+v", token)
	}
	if token.LastUsedAt == nil || *token.LastUsedAt != sampledAt.Add(10*time.Minute).Format(time.RFC3339) {
		t.Fatalf("unexpected last_used_at: %+v", token.LastUsedAt)
	}

	var hourlyUpload, hourlyDownload int64
	if err := db.db.QueryRowContext(ctx, `
		SELECT upload_bytes, download_bytes
		FROM traffic_user_hourly
		WHERE token_id = ?
	`, result.Token.ID).Scan(&hourlyUpload, &hourlyDownload); err != nil {
		t.Fatalf("query hourly rollup: %v", err)
	}
	if hourlyUpload != 85 || hourlyDownload != 80 {
		t.Fatalf("unexpected hourly rollup: upload=%d download=%d", hourlyUpload, hourlyDownload)
	}

	if _, err := db.RecordTrafficSamples(ctx, []RecordTrafficSampleInput{
		{
			SampledAt:        sampledAt,
			MetricType:       "outbound",
			MetricName:       "up_42",
			RawValueUpload:   50,
			RawValueDownload: 70,
		},
		{
			SampledAt:        sampledAt.Add(5 * time.Minute),
			MetricType:       "outbound",
			MetricName:       "up_42",
			RawValueUpload:   80,
			RawValueDownload: 120,
		},
	}); err != nil {
		t.Fatalf("record outbound samples: %v", err)
	}
	var outboundUpload, outboundDownload int64
	if err := db.db.QueryRowContext(ctx, `
		SELECT upload_bytes, download_bytes
		FROM traffic_outbound_daily
		WHERE outbound_tag = 'up_42'
	`).Scan(&outboundUpload, &outboundDownload); err != nil {
		t.Fatalf("query outbound rollup: %v", err)
	}
	if outboundUpload != 30 || outboundDownload != 50 {
		t.Fatalf("unexpected outbound rollup: upload=%d download=%d", outboundUpload, outboundDownload)
	}

	summaries, err := db.ListTokenTrafficAt(ctx, sampledAt)
	if err != nil {
		t.Fatalf("list token traffic: %v", err)
	}
	if len(summaries) != 1 || summaries[0].UsedTotalBytes != 165 || summaries[0].TodayTotalBytes != 165 || summaries[0].MonthTotalBytes != 165 || summaries[0].AuthUser != result.Account.AuthUser {
		t.Fatalf("unexpected traffic summary: %+v", summaries)
	}
}

func TestListTrafficDailyAtFillsRecentDays(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	team, err := db.CreateTeam(ctx, CreateTeamInput{Name: "Daily Team"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	user, err := db.CreateUser(ctx, CreateUserInput{TeamID: &team.ID, Name: "Daily User"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	result, err := db.CreateToken(ctx, "secret", "https://flux.example", CreateTokenInput{
		UserID:     user.ID,
		Name:       "daily token",
		ExpireDays: 30,
		QuotaBytes: 1000,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	now := time.Date(2026, 4, 29, 8, 0, 0, 0, time.UTC)
	previousDay := now.AddDate(0, 0, -1)
	for _, input := range []RecordTrafficSampleInput{
		{
			SampledAt:        previousDay,
			MetricType:       "user",
			MetricName:       result.Account.AuthUser,
			RawValueUpload:   0,
			RawValueDownload: 0,
		},
		{
			SampledAt:        previousDay.Add(time.Hour),
			MetricType:       "user",
			MetricName:       result.Account.AuthUser,
			RawValueUpload:   10,
			RawValueDownload: 15,
		},
		{
			SampledAt:        now,
			MetricType:       "user",
			MetricName:       result.Account.AuthUser,
			RawValueUpload:   40,
			RawValueDownload: 70,
		},
	} {
		if _, err := db.RecordTrafficSample(ctx, input); err != nil {
			t.Fatalf("record traffic sample: %v", err)
		}
	}

	days, err := db.ListTrafficDailyAt(ctx, now, 3)
	if err != nil {
		t.Fatalf("list traffic daily: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("expected 3 days, got %+v", days)
	}
	if days[0].Day != "2026-04-27" || days[0].TotalBytes != 0 {
		t.Fatalf("expected empty first day, got %+v", days[0])
	}
	if days[1].Day != "2026-04-28" || days[1].UploadBytes != 10 || days[1].DownloadBytes != 15 || days[1].TotalBytes != 25 {
		t.Fatalf("unexpected previous day summary: %+v", days[1])
	}
	if days[2].Day != "2026-04-29" || days[2].UploadBytes != 30 || days[2].DownloadBytes != 55 || days[2].TotalBytes != 85 {
		t.Fatalf("unexpected current day summary: %+v", days[2])
	}
}

func TestTrafficUsageMarksAndRestoresOverQuotaToken(t *testing.T) {
	ctx := context.Background()
	db := openTestStore(t)

	team, err := db.CreateTeam(ctx, CreateTeamInput{Name: "Quota Team"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	user, err := db.CreateUser(ctx, CreateUserInput{TeamID: &team.ID, Name: "Quota User"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	result, err := db.CreateToken(ctx, "secret", "https://flux.example", CreateTokenInput{
		UserID:     user.ID,
		Name:       "quota token",
		ExpireDays: 30,
		QuotaBytes: 100,
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	sampledAt := time.Date(2026, 4, 29, 6, 22, 0, 0, time.UTC)
	if _, err := db.RecordTrafficSample(ctx, RecordTrafficSampleInput{
		SampledAt:        sampledAt,
		MetricType:       "user",
		MetricName:       result.Account.AuthUser,
		RawValueUpload:   10,
		RawValueDownload: 20,
	}); err != nil {
		t.Fatalf("record baseline sample: %v", err)
	}
	if _, err := db.RecordTrafficSample(ctx, RecordTrafficSampleInput{
		SampledAt:        sampledAt.Add(time.Minute),
		MetricType:       "user",
		MetricName:       result.Account.AuthUser,
		RawValueUpload:   70,
		RawValueDownload: 70,
	}); err != nil {
		t.Fatalf("record over quota sample: %v", err)
	}

	token, err := db.GetToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get token: %v", err)
	}
	if token.Status != "over_quota" || token.UsedUploadBytes+token.UsedDownloadBytes != 110 {
		t.Fatalf("token should be over quota: %+v", token)
	}
	account, err := db.GetGatewayAccountByToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get gateway account: %v", err)
	}
	if account.Status != "over_quota" {
		t.Fatalf("gateway account should be over quota: %+v", account)
	}

	token, err = db.AddTokenQuota(ctx, result.Token.ID, 5)
	if err != nil {
		t.Fatalf("add insufficient quota: %v", err)
	}
	if token.Status != "over_quota" {
		t.Fatalf("token should stay over quota when added quota is insufficient: %+v", token)
	}
	token, err = db.AddTokenQuota(ctx, result.Token.ID, 100)
	if err != nil {
		t.Fatalf("add restoring quota: %v", err)
	}
	if token.Status != "active" {
		t.Fatalf("token should restore after enough quota: %+v", token)
	}
	account, err = db.GetGatewayAccountByToken(ctx, result.Token.ID)
	if err != nil {
		t.Fatalf("get restored gateway account: %v", err)
	}
	if account.Status != "active" {
		t.Fatalf("gateway account should restore after enough quota: %+v", account)
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

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range left {
		seen[value] = true
	}
	for _, value := range right {
		if !seen[value] {
			return false
		}
	}
	return true
}
