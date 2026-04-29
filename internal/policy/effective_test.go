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
	return store.Policy{
		ScopeType: scope,
		ScopeID:   &scopeID,
		MaxNodes:  maxNodes,
		Status:    "active",
	}
}
