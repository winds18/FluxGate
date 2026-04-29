package policy

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

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
	filtered := orderedEligibleVirtualNodes(virtualNodes)
	allowedValues, matchedAllowed := EffectiveAllowedVirtualNodes(token, policies)
	if matchedAllowed && len(allowedValues) > 0 {
		filtered = filterAllowedVirtualNodes(filtered, allowedValues)
	}
	maxNodes := EffectiveMaxNodes(token, policies)
	if maxNodes > 0 && len(filtered) > maxNodes {
		filtered = filtered[:maxNodes]
	}
	return filtered
}

func VirtualNodeAllowed(token store.TokenWithAccount, target store.VirtualNode, virtualNodes []store.VirtualNode, policies []store.Policy) bool {
	for _, node := range FilterVirtualNodes(virtualNodes, token, policies) {
		if sameVirtualNode(node, target) {
			return true
		}
	}
	return false
}

func EffectiveAllowedVirtualNodes(token store.TokenWithAccount, policies []store.Policy) ([]string, bool) {
	for _, scope := range []string{"token", "user", "team"} {
		values, matched := scopedAllowedVirtualNodes(token, policies, scope)
		if matched {
			return values, true
		}
	}
	return nil, false
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

func scopedAllowedVirtualNodes(token store.TokenWithAccount, policies []store.Policy, scope string) ([]string, bool) {
	matched := false
	seen := map[string]bool{}
	var values []string
	for _, item := range policies {
		if item.Status != "active" || item.ScopeType != scope || !scopeMatches(token, item) {
			continue
		}
		matched = true
		for _, value := range parseAllowedVirtualNodes(item.AllowedVirtualNodes) {
			key := strings.ToLower(value)
			if seen[key] {
				continue
			}
			seen[key] = true
			values = append(values, value)
		}
	}
	return values, matched
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

func filterAllowedVirtualNodes(virtualNodes []store.VirtualNode, allowedValues []string) []store.VirtualNode {
	filtered := make([]store.VirtualNode, 0, len(virtualNodes))
	for _, node := range virtualNodes {
		if virtualNodeAllowedByValues(node, allowedValues) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func virtualNodeAllowedByValues(node store.VirtualNode, allowedValues []string) bool {
	for _, value := range allowedValues {
		if strconv.FormatInt(node.ID, 10) == value || strings.EqualFold(node.Name, value) {
			return true
		}
	}
	return false
}

func parseAllowedVirtualNodes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var items []any
		if err := json.Unmarshal([]byte(raw), &items); err == nil {
			values := make([]string, 0, len(items))
			for _, item := range items {
				if value := virtualNodeValueString(item); value != "" {
					values = append(values, value)
				}
			}
			return values
		}
	}

	var values []string
	for _, item := range strings.Split(raw, ",") {
		if value := strings.TrimSpace(item); value != "" {
			values = append(values, value)
		}
	}
	return values
}

func virtualNodeValueString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	}
	return ""
}
