package store

import "context"

type CreateVirtualNodeInput struct {
	Name           string `json:"name"`
	ListenProtocol string `json:"listen_protocol"`
	ListenPort     int    `json:"listen_port"`
	TagSelector    string `json:"tag_selector"`
	Strategy       string `json:"strategy"`
}

func (s *Store) CreateVirtualNode(ctx context.Context, input CreateVirtualNodeInput) (VirtualNode, error) {
	if input.ListenProtocol == "" {
		input.ListenProtocol = "vless"
	}
	if input.TagSelector == "" {
		input.TagSelector = "{}"
	}
	if input.Strategy == "" {
		input.Strategy = "selector"
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO virtual_nodes(name, listen_protocol, listen_port, tag_selector, strategy)
		VALUES (?, ?, ?, ?, ?)
	`, input.Name, input.ListenProtocol, input.ListenPort, input.TagSelector, input.Strategy)
	if err != nil {
		return VirtualNode{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
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
