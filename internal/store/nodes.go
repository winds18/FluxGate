package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/winds18/FluxGate/internal/naming"
	"github.com/winds18/FluxGate/internal/region"
)

type ImportNodesInput struct {
	SourceID            int64  `json:"source_id"`
	Content             string `json:"content"`
	MarkMissingInactive bool   `json:"mark_missing_inactive"`
}

func (s *Store) ImportNodes(ctx context.Context, input ImportNodesInput) (ImportResult, error) {
	source, err := s.GetSource(ctx, input.SourceID)
	if err != nil {
		return ImportResult{}, err
	}

	var result ImportResult
	var seenHashes []string
	for _, line := range strings.Split(input.Content, "\n") {
		uri := strings.TrimSpace(line)
		if uri == "" || strings.HasPrefix(uri, "#") {
			result.Skipped++
			continue
		}
		rawName := naming.RawNameFromURI(uri)
		displayName := naming.DisplayName(source.DisplayPrefix, rawName)
		detectedRegion := region.Normalize("", rawName)
		protocol := naming.ProtocolFromURI(uri)
		server, port := naming.ServerFromURI(uri)
		uriHash := hashURI(uri)
		seenHashes = append(seenHashes, uriHash)

		existing, err := s.nodeBySourceHash(ctx, input.SourceID, uriHash)
		if err != nil && err != sql.ErrNoRows {
			return result, err
		}
		if err == nil {
			nodeRegion := region.Normalize(existing.Region, rawName)
			_, err = s.db.ExecContext(ctx, `
				UPDATE upstream_nodes
				SET raw_name = ?,
				    display_name = CASE WHEN name_mode = 'auto' THEN ? ELSE display_name END,
				    uri = ?,
				    protocol = ?,
				    server = ?,
				    server_port = ?,
				    region = ?,
				    status = 'active',
				    last_seen_at = CURRENT_TIMESTAMP,
				    updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, rawName, displayName, uri, protocol, server, port, nodeRegion, existing.ID)
			if err != nil {
				return result, err
			}
			if err := s.replaceNodeTags(ctx, existing.ID, source.DefaultTags); err != nil {
				return result, err
			}
			result.Updated++
			continue
		}

		insertResult, err := s.db.ExecContext(ctx, `
			INSERT INTO upstream_nodes(source_id, raw_name, display_name, name_mode, uri, uri_hash, protocol, server, server_port, region, status, last_seen_at)
			VALUES (?, ?, ?, 'auto', ?, ?, ?, ?, ?, ?, 'active', CURRENT_TIMESTAMP)
		`, input.SourceID, rawName, displayName, uri, uriHash, protocol, server, port, detectedRegion)
		if err != nil {
			return result, err
		}
		nodeID, err := insertResult.LastInsertId()
		if err != nil {
			return result, err
		}
		if err := s.replaceNodeTags(ctx, nodeID, source.DefaultTags); err != nil {
			return result, err
		}
		result.Imported++
	}

	if input.MarkMissingInactive && len(seenHashes) > 0 {
		inactivated, err := s.markMissingSourceNodesInactive(ctx, input.SourceID, seenHashes)
		if err != nil {
			return result, err
		}
		result.Inactivated = inactivated
	}

	if _, err := s.db.ExecContext(ctx, "UPDATE upstream_sources SET last_sync_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?", input.SourceID); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Store) ListNodes(ctx context.Context) ([]Node, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT n.id, n.source_id, s.name, n.raw_name, n.display_name, n.name_mode, n.uri, n.uri_hash,
		       n.protocol, n.server, n.server_port, n.region, n.status, n.last_seen_at,
		       n.last_checked_at, n.last_error, n.created_at, n.updated_at,
		       COALESCE(GROUP_CONCAT(t.name), '') AS tags
		FROM upstream_nodes n
		JOIN upstream_sources s ON s.id = n.source_id
		LEFT JOIN upstream_node_tags nt ON nt.node_id = n.id
		LEFT JOIN tags t ON t.id = nt.tag_id
		GROUP BY n.id
		ORDER BY n.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		node, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *Store) GetNode(ctx context.Context, id int64) (Node, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT n.id, n.source_id, s.name, n.raw_name, n.display_name, n.name_mode, n.uri, n.uri_hash,
		       n.protocol, n.server, n.server_port, n.region, n.status, n.last_seen_at,
		       n.last_checked_at, n.last_error, n.created_at, n.updated_at,
		       COALESCE(GROUP_CONCAT(t.name), '') AS tags
		FROM upstream_nodes n
		JOIN upstream_sources s ON s.id = n.source_id
		LEFT JOIN upstream_node_tags nt ON nt.node_id = n.id
		LEFT JOIN tags t ON t.id = nt.tag_id
		WHERE n.id = ?
		GROUP BY n.id
	`, id)
	return scanNode(row)
}

func (s *Store) UpdateNodeDisplayName(ctx context.Context, id int64, displayName string) (Node, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = "Unnamed"
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE upstream_nodes
		SET display_name = ?, name_mode = 'manual', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, displayName, id); err != nil {
		return Node{}, err
	}
	return s.GetNode(ctx, id)
}

func (s *Store) ResetNodeDisplayName(ctx context.Context, id int64) (Node, error) {
	node, err := s.GetNode(ctx, id)
	if err != nil {
		return Node{}, err
	}
	source, err := s.GetSource(ctx, node.SourceID)
	if err != nil {
		return Node{}, err
	}
	displayName := naming.DisplayName(source.DisplayPrefix, node.RawName)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE upstream_nodes
		SET display_name = ?, name_mode = 'auto', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, displayName, id); err != nil {
		return Node{}, err
	}
	return s.GetNode(ctx, id)
}

func (s *Store) nodeBySourceHash(ctx context.Context, sourceID int64, uriHash string) (Node, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT n.id, n.source_id, s.name, n.raw_name, n.display_name, n.name_mode, n.uri, n.uri_hash,
		       n.protocol, n.server, n.server_port, n.region, n.status, n.last_seen_at,
		       n.last_checked_at, n.last_error, n.created_at, n.updated_at,
		       COALESCE(GROUP_CONCAT(t.name), '') AS tags
		FROM upstream_nodes n
		JOIN upstream_sources s ON s.id = n.source_id
		LEFT JOIN upstream_node_tags nt ON nt.node_id = n.id
		LEFT JOIN tags t ON t.id = nt.tag_id
		WHERE n.source_id = ? AND n.uri_hash = ?
		GROUP BY n.id
	`, sourceID, uriHash)
	return scanNode(row)
}

func (s *Store) markMissingSourceNodesInactive(ctx context.Context, sourceID int64, seenHashes []string) (int, error) {
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(seenHashes)), ",")
	args := make([]any, 0, len(seenHashes)+1)
	args = append(args, sourceID)
	for _, hash := range seenHashes {
		args = append(args, hash)
	}
	result, err := s.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE upstream_nodes
		SET status = 'inactive',
		    updated_at = CURRENT_TIMESTAMP
		WHERE source_id = ?
		  AND status = 'active'
		  AND uri_hash NOT IN (%s)
	`, placeholders), args...)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

type nodeScanner interface {
	Scan(dest ...any) error
}

func scanNode(scanner nodeScanner) (Node, error) {
	var node Node
	var tagCSV sql.NullString
	err := scanner.Scan(
		&node.ID,
		&node.SourceID,
		&node.SourceName,
		&node.RawName,
		&node.DisplayName,
		&node.NameMode,
		&node.URI,
		&node.URIHash,
		&node.Protocol,
		&node.Server,
		&node.ServerPort,
		&node.Region,
		&node.Status,
		&node.LastSeenAt,
		&node.LastCheckedAt,
		&node.LastError,
		&node.CreatedAt,
		&node.UpdatedAt,
		&tagCSV,
	)
	node.Tags = splitTagCSV(tagCSV.String)
	node.Region = region.Normalize(node.Region, node.RawName)
	return node, err
}

func (s *Store) replaceNodeTags(ctx context.Context, nodeID int64, rawTags string) error {
	if _, err := s.db.ExecContext(ctx, "DELETE FROM upstream_node_tags WHERE node_id = ?", nodeID); err != nil {
		return err
	}
	for _, name := range parseTagValues(rawTags) {
		tagID, err := s.ensureTag(ctx, name)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO upstream_node_tags(node_id, tag_id) VALUES (?, ?)", nodeID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureTag(ctx context.Context, name string) (int64, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("tag name is required")
	}
	if _, err := s.db.ExecContext(ctx, "INSERT OR IGNORE INTO tags(name) VALUES (?)", name); err != nil {
		return 0, err
	}
	var id int64
	err := s.db.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ?", name).Scan(&id)
	return id, err
}

func parseTagValues(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var values []string
		if err := json.Unmarshal([]byte(raw), &values); err == nil {
			return cleanTagValues(values)
		}
	}
	return cleanTagValues(strings.Split(raw, ","))
}

func cleanTagValues(values []string) []string {
	seen := map[string]bool{}
	cleaned := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if seen[key] {
			continue
		}
		seen[key] = true
		cleaned = append(cleaned, value)
	}
	return cleaned
}

func splitTagCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	return cleanTagValues(strings.Split(value, ","))
}

func hashURI(uri string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(uri)))
	return hex.EncodeToString(sum[:])
}
