package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type CreatePolicyInput struct {
	Name                string `json:"name"`
	ScopeType           string `json:"scope_type"`
	ScopeID             *int64 `json:"scope_id"`
	IncludeTags         string `json:"include_tags"`
	ExcludeTags         string `json:"exclude_tags"`
	AllowedVirtualNodes string `json:"allowed_virtual_nodes"`
	MaxNodes            int64  `json:"max_nodes"`
}

func (s *Store) CreatePolicy(ctx context.Context, input CreatePolicyInput) (Policy, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return Policy{}, fmt.Errorf("policy name is required")
	}
	input.ScopeType = strings.TrimSpace(input.ScopeType)
	if input.ScopeType == "" {
		input.ScopeType = "team"
	}
	if input.ScopeType != "team" && input.ScopeType != "user" && input.ScopeType != "token" {
		return Policy{}, fmt.Errorf("scope_type must be team, user, or token")
	}
	if input.MaxNodes < 0 {
		return Policy{}, fmt.Errorf("max_nodes must be greater than or equal to 0")
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO policies(name, scope_type, scope_id, include_tags, exclude_tags, allowed_virtual_nodes, max_nodes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.ScopeType, input.ScopeID, jsonArray(input.IncludeTags), jsonArray(input.ExcludeTags), jsonArray(input.AllowedVirtualNodes), input.MaxNodes)
	if err != nil {
		return Policy{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Policy{}, err
	}
	return s.GetPolicy(ctx, id)
}

func (s *Store) GetPolicy(ctx context.Context, id int64) (Policy, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, scope_type, scope_id, include_tags, exclude_tags, allowed_virtual_nodes, max_nodes, status, created_at, updated_at
		FROM policies
		WHERE id = ?
	`, id)
	return scanPolicy(row)
}

func (s *Store) ListPolicies(ctx context.Context) ([]Policy, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, scope_type, scope_id, include_tags, exclude_tags, allowed_virtual_nodes, max_nodes, status, created_at, updated_at
		FROM policies
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}
	return policies, rows.Err()
}

func jsonArray(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "[]"
	}
	return value
}

func scanPolicy(scanner scanner) (Policy, error) {
	var policy Policy
	var scopeID sql.NullInt64
	err := scanner.Scan(
		&policy.ID,
		&policy.Name,
		&policy.ScopeType,
		&scopeID,
		&policy.IncludeTags,
		&policy.ExcludeTags,
		&policy.AllowedVirtualNodes,
		&policy.MaxNodes,
		&policy.Status,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return Policy{}, err
	}
	if scopeID.Valid {
		policy.ScopeID = &scopeID.Int64
	}
	return policy, nil
}
