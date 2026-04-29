package policy

import (
	"sort"

	"github.com/winds18/FluxGate/internal/store"
)

func EffectiveMaxNodes(token store.TokenWithAccount, policies []store.Policy) int {
	for _, scope := range []string{"token", "user", "team"} {
		maxNodes, matched := scopedMaxNodes(token, policies, scope)
		if matched {
			return maxNodes
		}
	}
	return 0
}

func FilterVirtualNodes(virtualNodes []store.VirtualNode, token store.TokenWithAccount, policies []store.Policy) []store.VirtualNode {
	maxNodes := EffectiveMaxNodes(token, policies)
	if maxNodes <= 0 {
		return virtualNodes
	}
	filtered := make([]store.VirtualNode, 0, len(virtualNodes))
	for index, node := range orderedEligibleVirtualNodes(virtualNodes) {
		if index >= maxNodes {
			break
		}
		filtered = append(filtered, node)
	}
	return filtered
}

func VirtualNodeAllowed(token store.TokenWithAccount, target store.VirtualNode, virtualNodes []store.VirtualNode, policies []store.Policy) bool {
	maxNodes := EffectiveMaxNodes(token, policies)
	if maxNodes <= 0 {
		return true
	}
	for index, node := range orderedEligibleVirtualNodes(virtualNodes) {
		if sameVirtualNode(node, target) {
			return index < maxNodes
		}
	}
	return false
}

func scopedMaxNodes(token store.TokenWithAccount, policies []store.Policy, scope string) (int, bool) {
	matched := false
	maxNodes := 0
	for _, item := range policies {
		if item.Status != "active" || item.ScopeType != scope || !scopeMatches(token, item) {
			continue
		}
		matched = true
		if item.MaxNodes <= 0 {
			continue
		}
		value := int(item.MaxNodes)
		if maxNodes == 0 || value < maxNodes {
			maxNodes = value
		}
	}
	return maxNodes, matched
}

func scopeMatches(token store.TokenWithAccount, item store.Policy) bool {
	if item.ScopeID == nil {
		return false
	}
	switch item.ScopeType {
	case "token":
		return *item.ScopeID == token.ID
	case "user":
		return *item.ScopeID == token.UserID
	case "team":
		return token.UserTeamID != nil && *item.ScopeID == *token.UserTeamID
	default:
		return false
	}
}

func sameVirtualNode(left store.VirtualNode, right store.VirtualNode) bool {
	if left.ID > 0 && right.ID > 0 {
		return left.ID == right.ID
	}
	return left.Name == right.Name && left.ListenPort == right.ListenPort && left.ListenProtocol == right.ListenProtocol
}

func orderedEligibleVirtualNodes(virtualNodes []store.VirtualNode) []store.VirtualNode {
	eligible := make([]store.VirtualNode, 0, len(virtualNodes))
	for _, node := range virtualNodes {
		if node.Status == "active" && node.ListenProtocol == "vless" {
			eligible = append(eligible, node)
		}
	}
	sort.SliceStable(eligible, func(i, j int) bool {
		if eligible[i].ID != eligible[j].ID {
			return eligible[i].ID < eligible[j].ID
		}
		if eligible[i].ListenPort != eligible[j].ListenPort {
			return eligible[i].ListenPort < eligible[j].ListenPort
		}
		return eligible[i].Name < eligible[j].Name
	})
	return eligible
}
