package singbox

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type CheckResult struct {
	Valid                 bool     `json:"valid"`
	ConfigHash            string   `json:"config_hash"`
	InboundCount          int      `json:"inbound_count"`
	OutboundCount         int      `json:"outbound_count"`
	UpstreamOutboundCount int      `json:"upstream_outbound_count"`
	UserCount             int      `json:"user_count"`
	Messages              []string `json:"messages"`
}

func CheckConfig(config Config) (CheckResult, error) {
	body, err := Marshal(config)
	if err != nil {
		return CheckResult{}, err
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return CheckResult{}, err
	}

	result := CheckResult{
		Valid:                 true,
		ConfigHash:            hashConfig(body),
		InboundCount:          len(config.Inbounds),
		OutboundCount:         len(config.Outbounds),
		UpstreamOutboundCount: countUpstreamOutbounds(config.Outbounds),
		UserCount:             countInboundUsers(config.Inbounds),
	}
	result.Messages = append(result.Messages, validateConfigShape(config)...)
	if len(result.Messages) > 0 {
		result.Valid = false
	}
	return result, nil
}

func validateConfigShape(config Config) []string {
	var messages []string
	if config.Log == nil {
		messages = append(messages, "missing log section")
	}
	if len(config.Outbounds) == 0 {
		messages = append(messages, "missing outbounds")
	}
	if config.Route == nil {
		messages = append(messages, "missing route section")
	}
	for index, inbound := range config.Inbounds {
		if strings.TrimSpace(inbound.Type) == "" {
			messages = append(messages, fmt.Sprintf("inbound %d missing type", index))
		}
		if strings.TrimSpace(inbound.Tag) == "" {
			messages = append(messages, fmt.Sprintf("inbound %d missing tag", index))
		}
		if inbound.ListenPort <= 0 || inbound.ListenPort > 65535 {
			messages = append(messages, fmt.Sprintf("inbound %s has invalid listen port", inbound.Tag))
		}
	}
	for index, outbound := range config.Outbounds {
		if strings.TrimSpace(stringFromAny(outbound["type"])) == "" {
			messages = append(messages, fmt.Sprintf("outbound %d missing type", index))
		}
		if strings.TrimSpace(stringFromAny(outbound["tag"])) == "" {
			messages = append(messages, fmt.Sprintf("outbound %d missing tag", index))
		}
	}
	return messages
}

func hashConfig(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func countInboundUsers(inbounds []Inbound) int {
	total := 0
	for _, inbound := range inbounds {
		total += len(inbound.Users)
	}
	return total
}

func countUpstreamOutbounds(outbounds []map[string]any) int {
	total := 0
	for _, outbound := range outbounds {
		tag := stringFromAny(outbound["tag"])
		if strings.HasPrefix(tag, "up_") {
			total++
		}
	}
	return total
}
