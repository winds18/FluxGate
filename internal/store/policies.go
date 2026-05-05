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

type UpdatePolicyInput struct {
	Name                *string `json:"name"`
	ScopeType           *string `json:"scope_type"`
	ScopeID             *int64  `json:"scope_id"`
	IncludeTags         *string `json:"include_tags"`
	ExcludeTags         *string `json:"exclude_tags"`
	AllowedVirtualNodes *string `json:"allowed_virtual_nodes"`
	MaxNodes            *int64  `json:"max_nodes"`
	Status              *string `json:"status"`
}

func (s *Store) CreatePolicy(ctx context.Context, input CreatePolicyInput) (Policy, error) {
	fields, err := normalizePolicyFields(policyFields{
		Name:                input.Name,
		ScopeType:           input.ScopeType,
		ScopeID:             input.ScopeID,
		IncludeTags:         input.IncludeTags,
		ExcludeTags:         input.ExcludeTags,
		AllowedVirtualNodes: input.AllowedVirtualNodes,
		MaxNodes:            input.MaxNodes,
		Status:              "active",
	})
	if err != nil {
		return Policy{}, err
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO policies(name, scope_type, scope_id, include_tags, exclude_tags, allowed_virtual_nodes, max_nodes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, fields.Name, fields.ScopeType, fields.ScopeID, fields.IncludeTags, fields.ExcludeTags, fields.AllowedVirtualNodes, fields.MaxNodes, fields.Status)
	if err != nil {
		return Policy{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Policy{}, err
	}
	return s.GetPolicy(ctx, id)
}

func (s *Store) UpdatePolicy(ctx context.Context, id int64, input UpdatePolicyInput) (Policy, error) {
	current, err := s.GetPolicy(ctx, id)
	if err != nil {
		return Policy{}, err
	}
	fields := policyFields{
		Name:                current.Name,
		ScopeType:           current.ScopeType,
		ScopeID:             current.ScopeID,
		IncludeTags:         current.IncludeTags,
		ExcludeTags:         current.ExcludeTags,
		AllowedVirtualNodes: current.AllowedVirtualNodes,
		MaxNodes:            current.MaxNodes,
		Status:              current.Status,
	}
	if input.Name != nil {
		fields.Name = *input.Name
	}
	if input.ScopeType != nil {
		fields.ScopeType = *input.ScopeType
	}
	if input.ScopeID != nil {
		fields.ScopeID = input.ScopeID
	}
	if input.IncludeTags != nil {
		fields.IncludeTags = *input.IncludeTags
	}
	if input.ExcludeTags != nil {
		fields.ExcludeTags = *input.ExcludeTags
	}
	if input.AllowedVirtualNodes != nil {
		fields.AllowedVirtualNodes = *input.AllowedVirtualNodes
	}
	if input.MaxNodes != nil {
		fields.MaxNodes = *input.MaxNodes
	}
	if input.Status != nil {
		fields.Status = *input.Status
	}

	fields, err = normalizePolicyFields(fields)
	if err != nil {
		return Policy{}, err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE policies
		SET name = ?, scope_type = ?, scope_id = ?, include_tags = ?, exclude_tags = ?, allowed_virtual_nodes = ?, max_nodes = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, fields.Name, fields.ScopeType, fields.ScopeID, fields.IncludeTags, fields.ExcludeTags, fields.AllowedVirtualNodes, fields.MaxNodes, fields.Status, id)
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

type policyFields struct {
	Name                string
	ScopeType           string
	ScopeID             *int64
	IncludeTags         string
	ExcludeTags         string
	AllowedVirtualNodes string
	MaxNodes            int64
	Status              string
}

func normalizePolicyFields(fields policyFields) (policyFields, error) {
	fields.Name = strings.TrimSpace(fields.Name)
	if fields.Name == "" {
		return policyFields{}, fmt.Errorf("policy name is required")
	}
	fields.ScopeType = strings.TrimSpace(fields.ScopeType)
	if fields.ScopeType == "" {
		fields.ScopeType = "team"
	}
	if fields.ScopeType != "team" && fields.ScopeType != "user" && fields.ScopeType != "token" {
		return policyFields{}, fmt.Errorf("scope_type must be team, user, or token")
	}
	if fields.ScopeID != nil && *fields.ScopeID <= 0 {
		return policyFields{}, fmt.Errorf("scope_id must be greater than 0")
	}
	fields.IncludeTags = jsonArray(fields.IncludeTags)
	fields.ExcludeTags = jsonArray(fields.ExcludeTags)
	fields.AllowedVirtualNodes = jsonArray(fields.AllowedVirtualNodes)
	if fields.MaxNodes < 0 {
		return policyFields{}, fmt.Errorf("max_nodes must be greater than or equal to 0")
	}
	fields.Status = strings.TrimSpace(fields.Status)
	if fields.Status == "" {
		fields.Status = "active"
	}
	if fields.Status != "active" && fields.Status != "inactive" {
		return policyFields{}, fmt.Errorf("status must be active or inactive")
	}
	return fields, nil
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
