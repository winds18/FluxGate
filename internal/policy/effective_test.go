package policy

import (
	"testing"

	"github.com/winds18/FluxGate/internal/store"
)

func TestEffectiveMaxNodesUsesHighestScope(t *testing.T) {
	teamID := int64(10)
	userID := int64(20)
	tokenID := int64(30)
	token := store.TokenWithAccount{
		Token:      store.Token{ID: tokenID, UserID: userID},
		UserTeamID: &teamID,
	}
	policies := []store.Policy{
		policy("team", teamID, 1),
		policy("user", userID, 2),
		policy("token", tokenID, 3),
	}

	if got := EffectiveMaxNodes(token, policies); got != 3 {
		t.Fatalf("expected token scope max_nodes to win, got %d", got)
	}
}

func TestFilterVirtualNodesAppliesTeamMaxNodes(t *testing.T) {
	teamID := int64(10)
	token := store.TokenWithAccount{Token: store.Token{ID: 30, UserID: 20}, UserTeamID: &teamID}
	nodes := []store.VirtualNode{
		{ID: 3, Name: "FluxGate-JP", ListenProtocol: "vless", ListenPort: 8445, Status: "active"},
		{ID: 2, Name: "FluxGate-SG", ListenProtocol: "vless", ListenPort: 8444, Status: "active"},
		{ID: 1, Name: "FluxGate-HK", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}

	filtered := FilterVirtualNodes(nodes, token, []store.Policy{policy("team", teamID, 2)})
	if len(filtered) != 2 || filtered[0].Name != "FluxGate-HK" || filtered[1].Name != "FluxGate-SG" {
		t.Fatalf("unexpected filtered nodes: %+v", filtered)
	}
	if !VirtualNodeAllowed(token, nodes[1], nodes, []store.Policy{policy("team", teamID, 2)}) {
		t.Fatalf("second node should be allowed")
	}
	if VirtualNodeAllowed(token, nodes[0], nodes, []store.Policy{policy("team", teamID, 2)}) {
		t.Fatalf("third node should be blocked")
	}
}

func TestFilterVirtualNodesAppliesAllowedVirtualNodes(t *testing.T) {
	teamID := int64(10)
	token := store.TokenWithAccount{Token: store.Token{ID: 30, UserID: 20}, UserTeamID: &teamID}
	nodes := []store.VirtualNode{
		{ID: 3, Name: "FluxGate-JP", ListenProtocol: "vless", ListenPort: 8445, Status: "active"},
		{ID: 2, Name: "FluxGate-SG", ListenProtocol: "vless", ListenPort: 8444, Status: "active"},
		{ID: 1, Name: "FluxGate-HK", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
	}

	filtered := FilterVirtualNodes(nodes, token, []store.Policy{policyWithAllowed("team", teamID, `["fluxgate-hk",2]`, 0)})
	if len(filtered) != 2 || filtered[0].Name != "FluxGate-HK" || filtered[1].Name != "FluxGate-SG" {
		t.Fatalf("unexpected allowed virtual nodes: %+v", filtered)
	}
	if VirtualNodeAllowed(token, nodes[0], nodes, []store.Policy{policyWithAllowed("team", teamID, `["FluxGate-HK"]`, 0)}) {
		t.Fatalf("node outside allowed virtual nodes should be blocked")
	}
}

func TestTokenScopeEmptyAllowedVirtualNodesOverridesTeamWhitelist(t *testing.T) {
	teamID := int64(10)
	tokenID := int64(30)
	token := store.TokenWithAccount{Token: store.Token{ID: tokenID, UserID: 20}, UserTeamID: &teamID}
	nodes := []store.VirtualNode{
		{ID: 1, Name: "FluxGate-HK", ListenProtocol: "vless", ListenPort: 8443, Status: "active"},
		{ID: 2, Name: "FluxGate-SG", ListenProtocol: "vless", ListenPort: 8444, Status: "active"},
	}
	policies := []store.Policy{
		policyWithAllowed("team", teamID, `["FluxGate-HK"]`, 0),
		policyWithAllowed("token", tokenID, `[]`, 0),
	}

	filtered := FilterVirtualNodes(nodes, token, policies)
	if len(filtered) != 2 {
		t.Fatalf("token-scope empty allowed_virtual_nodes should leave nodes unrestricted, got %+v", filtered)
	}
}

func TestEffectiveTagSelectorUsesHighestScope(t *testing.T) {
	teamID := int64(10)
	tokenID := int64(30)
	token := store.TokenWithAccount{Token: store.Token{ID: tokenID, UserID: 20}, UserTeamID: &teamID}
	policies := []store.Policy{
		policyWithTags("team", teamID, "HK", "Backup"),
		policyWithTags("token", tokenID, "SG, Premium", "Slow"),
	}

	selector, matched := EffectiveTagSelector(token, policies)
	if !matched {
		t.Fatal("expected tag selector policy to match")
	}
	if len(selector.Include) != 2 || selector.Include[0] != "SG" || selector.Include[1] != "Premium" {
		t.Fatalf("token scope include_tags should win, got %+v", selector.Include)
	}
	if len(selector.Exclude) != 1 || selector.Exclude[0] != "Slow" {
		t.Fatalf("token scope exclude_tags should win, got %+v", selector.Exclude)
	}
}

func TestEffectiveTagSelectorEmptyHigherScopeOverridesLowerScope(t *testing.T) {
	teamID := int64(10)
	tokenID := int64(30)
	token := store.TokenWithAccount{Token: store.Token{ID: tokenID, UserID: 20}, UserTeamID: &teamID}
	policies := []store.Policy{
		policyWithTags("team", teamID, "HK", ""),
		policyWithTags("token", tokenID, "[]", "[]"),
	}

	selector, matched := EffectiveTagSelector(token, policies)
	if !matched {
		t.Fatal("expected token scope policy to match")
	}
	if len(selector.Include) != 0 || len(selector.Exclude) != 0 {
		t.Fatalf("empty token scope tag policy should be unrestricted, got %+v", selector)
	}
}

func TestInactivePolicyDoesNotLimitNodes(t *testing.T) {
	teamID := int64(10)
	token := store.TokenWithAccount{Token: store.Token{ID: 30, UserID: 20}, UserTeamID: &teamID}
	item := policy("team", teamID, 1)
	item.Status = "inactive"

	if got := EffectiveMaxNodes(token, []store.Policy{item}); got != 0 {
		t.Fatalf("inactive policy should not limit nodes, got %d", got)
	}
}

func policy(scope string, scopeID int64, maxNodes int64) store.Policy {
	return policyWithAllowed(scope, scopeID, "", maxNodes)
}

func policyWithAllowed(scope string, scopeID int64, allowedVirtualNodes string, maxNodes int64) store.Policy {
	return store.Policy{
		ScopeType:           scope,
		ScopeID:             &scopeID,
		AllowedVirtualNodes: allowedVirtualNodes,
		MaxNodes:            maxNodes,
		Status:              "active",
	}
}

func policyWithTags(scope string, scopeID int64, includeTags string, excludeTags string) store.Policy {
	return store.Policy{
		ScopeType:   scope,
		ScopeID:     &scopeID,
		IncludeTags: includeTags,
		ExcludeTags: excludeTags,
		Status:      "active",
	}
}
