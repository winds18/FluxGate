package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type CreateVirtualNodeInput struct {
	Name           string `json:"name"`
	ListenProtocol string `json:"listen_protocol"`
	ListenPort     int    `json:"listen_port"`
	TagSelector    string `json:"tag_selector"`
	Strategy       string `json:"strategy"`
}

type UpdateVirtualNodeInput struct {
	Name           *string `json:"name"`
	ListenProtocol *string `json:"listen_protocol"`
	ListenPort     *int    `json:"listen_port"`
	TagSelector    *string `json:"tag_selector"`
	Strategy       *string `json:"strategy"`
	Status         *string `json:"status"`
}

func (s *Store) CreateVirtualNode(ctx context.Context, input CreateVirtualNodeInput) (VirtualNode, error) {
	name, listenProtocol, listenPort, tagSelector, strategy, status, err := normalizeVirtualNodeFields(
		input.Name,
		input.ListenProtocol,
		input.ListenPort,
		input.TagSelector,
		input.Strategy,
		"active",
	)
	if err != nil {
		return VirtualNode{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO virtual_nodes(name, listen_protocol, listen_port, tag_selector, strategy, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, name, listenProtocol, listenPort, tagSelector, strategy, status)
	if err != nil {
		return VirtualNode{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return VirtualNode{}, err
	}
	return s.GetVirtualNode(ctx, id)
}

func (s *Store) UpdateVirtualNode(ctx context.Context, id int64, input UpdateVirtualNodeInput) (VirtualNode, error) {
	node, err := s.GetVirtualNode(ctx, id)
	if err != nil {
		return VirtualNode{}, err
	}
	name := node.Name
	if input.Name != nil {
		name = *input.Name
	}
	listenProtocol := node.ListenProtocol
	if input.ListenProtocol != nil {
		listenProtocol = *input.ListenProtocol
	}
	listenPort := node.ListenPort
	if input.ListenPort != nil {
		listenPort = *input.ListenPort
	}
	tagSelector := node.TagSelector
	if input.TagSelector != nil {
		tagSelector = *input.TagSelector
	}
	strategy := node.Strategy
	if input.Strategy != nil {
		strategy = *input.Strategy
	}
	status := node.Status
	if input.Status != nil {
		status = *input.Status
	}

	name, listenProtocol, listenPort, tagSelector, strategy, status, err = normalizeVirtualNodeFields(
		name,
		listenProtocol,
		listenPort,
		tagSelector,
		strategy,
		status,
	)
	if err != nil {
		return VirtualNode{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE virtual_nodes
		SET name = ?,
		    listen_protocol = ?,
		    listen_port = ?,
		    tag_selector = ?,
		    strategy = ?,
		    status = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, name, listenProtocol, listenPort, tagSelector, strategy, status, id); err != nil {
		return VirtualNode{}, err
	}
	return s.GetVirtualNode(ctx, id)
}

func (s *Store) GetVirtualNode(ctx context.Context, id int64) (VirtualNode, error) {
	var node VirtualNode
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, listen_protocol, listen_port, tag_selector, strategy, status, created_at, updated_at
		FROM virtual_nodes
		WHERE id = ?
	`, id).Scan(&node.ID, &node.Name, &node.ListenProtocol, &node.ListenPort, &node.TagSelector, &node.Strategy, &node.Status, &node.CreatedAt, &node.UpdatedAt)
	return node, err
}

func (s *Store) ListVirtualNodes(ctx context.Context) ([]VirtualNode, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, listen_protocol, listen_port, tag_selector, strategy, status, created_at, updated_at
		FROM virtual_nodes
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var nodes []VirtualNode
	for rows.Next() {
		var node VirtualNode
		if err := rows.Scan(&node.ID, &node.Name, &node.ListenProtocol, &node.ListenPort, &node.TagSelector, &node.Strategy, &node.Status, &node.CreatedAt, &node.UpdatedAt); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func normalizeVirtualNodeFields(name, listenProtocol string, listenPort int, tagSelector, strategy, status string) (string, string, int, string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", 0, "", "", "", fmt.Errorf("virtual node name is required")
	}
	listenProtocol = strings.TrimSpace(listenProtocol)
	if listenProtocol == "" {
		listenProtocol = "vless"
	}
	if listenProtocol != "vless" {
		return "", "", 0, "", "", "", fmt.Errorf("listen_protocol must be vless")
	}
	if listenPort <= 0 || listenPort > 65535 {
		return "", "", 0, "", "", "", fmt.Errorf("listen_port must be between 1 and 65535")
	}
	tagSelector = strings.TrimSpace(tagSelector)
	if tagSelector == "" {
		tagSelector = "{}"
	}
	if !json.Valid([]byte(tagSelector)) {
		return "", "", 0, "", "", "", fmt.Errorf("tag_selector must be valid JSON")
	}
	strategy = strings.TrimSpace(strategy)
	if strategy == "" {
		strategy = "selector"
	}
	if strategy != "selector" {
		return "", "", 0, "", "", "", fmt.Errorf("strategy must be selector")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "active"
	}
	if status != "active" && status != "inactive" {
		return "", "", 0, "", "", "", fmt.Errorf("status must be active or inactive")
	}
	return name, listenProtocol, listenPort, tagSelector, strategy, status, nil
}
