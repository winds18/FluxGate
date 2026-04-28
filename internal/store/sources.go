package store

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/winds18/FluxGate/internal/naming"
)

var whitespaceRE = regexp.MustCompile(`\s+`)

type CreateSourceInput struct {
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	URL                    string `json:"url"`
	RawContent             string `json:"raw_content"`
	DisplayPrefix          string `json:"display_prefix"`
	DefaultTags            string `json:"default_tags"`
	RefreshIntervalMinutes int64  `json:"refresh_interval_minutes"`
}

func (s *Store) CreateSource(ctx context.Context, input CreateSourceInput) (Source, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "Unnamed Source"
	}
	sourceType := strings.TrimSpace(input.Type)
	if sourceType == "" {
		sourceType = "manual"
	}
	defaultTags := strings.TrimSpace(input.DefaultTags)
	if defaultTags == "" {
		defaultTags = "[]"
	}

	prefixMode := "manual"
	prefix := naming.NormalizeManualPrefix(input.DisplayPrefix)
	if prefix == "" {
		prefixMode = "auto"
		var err error
		prefix, err = s.uniqueAutoPrefix(ctx, name, input.URL, 0)
		if err != nil {
			return Source{}, err
		}
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO upstream_sources(name, type, url, raw_content, prefix_mode, display_prefix, default_tags, refresh_interval_minutes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, name, sourceType, strings.TrimSpace(input.URL), input.RawContent, prefixMode, prefix, defaultTags, input.RefreshIntervalMinutes)
	if err != nil {
		return Source{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Source{}, err
	}
	return s.GetSource(ctx, id)
}

func (s *Store) GetSource(ctx context.Context, id int64) (Source, error) {
	var source Source
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, type, url, raw_content, prefix_mode, display_prefix, default_tags,
		       refresh_interval_minutes, status, last_sync_at, last_error, created_at, updated_at
		FROM upstream_sources
		WHERE id = ?
	`, id).Scan(
		&source.ID,
		&source.Name,
		&source.Type,
		&source.URL,
		&source.RawContent,
		&source.PrefixMode,
		&source.DisplayPrefix,
		&source.DefaultTags,
		&source.RefreshIntervalMinutes,
		&source.Status,
		&source.LastSyncAt,
		&source.LastError,
		&source.CreatedAt,
		&source.UpdatedAt,
	)
	return source, err
}

func (s *Store) ListSources(ctx context.Context) ([]Source, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, url, raw_content, prefix_mode, display_prefix, default_tags,
		       refresh_interval_minutes, status, last_sync_at, last_error, created_at, updated_at
		FROM upstream_sources
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []Source
	for rows.Next() {
		var source Source
		if err := rows.Scan(
			&source.ID,
			&source.Name,
			&source.Type,
			&source.URL,
			&source.RawContent,
			&source.PrefixMode,
			&source.DisplayPrefix,
			&source.DefaultTags,
			&source.RefreshIntervalMinutes,
			&source.Status,
			&source.LastSyncAt,
			&source.LastError,
			&source.CreatedAt,
			&source.UpdatedAt,
		); err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (s *Store) UpdateSourcePrefix(ctx context.Context, id int64, prefix string) (Source, error) {
	prefix = naming.NormalizeManualPrefix(prefix)
	if prefix == "" {
		source, err := s.GetSource(ctx, id)
		if err != nil {
			return Source{}, err
		}
		prefix, err = s.uniqueAutoPrefix(ctx, source.Name, source.URL, id)
		if err != nil {
			return Source{}, err
		}
		if _, err := s.db.ExecContext(ctx, `
			UPDATE upstream_sources
			SET prefix_mode = 'auto', display_prefix = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, prefix, id); err != nil {
			return Source{}, err
		}
	} else {
		if _, err := s.db.ExecContext(ctx, `
			UPDATE upstream_sources
			SET prefix_mode = 'manual', display_prefix = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, prefix, id); err != nil {
			return Source{}, err
		}
	}
	if err := s.RegenerateSourceNodeNames(ctx, id); err != nil {
		return Source{}, err
	}
	return s.GetSource(ctx, id)
}

func (s *Store) RegenerateSourceNodeNames(ctx context.Context, sourceID int64) error {
	source, err := s.GetSource(ctx, sourceID)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE upstream_nodes
		SET display_name = ? || raw_name,
		    updated_at = CURRENT_TIMESTAMP
		WHERE source_id = ?
		  AND name_mode = 'auto'
	`, source.DisplayPrefix, sourceID)
	return err
}

func (s *Store) UpdateSourceRawContent(ctx context.Context, id int64, rawContent string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE upstream_sources
		SET raw_content = ?,
		    last_error = '',
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, rawContent, id)
	return err
}

func (s *Store) SetSourceSyncError(ctx context.Context, id int64, message string) error {
	message = whitespaceRE.ReplaceAllString(strings.TrimSpace(message), " ")
	if len(message) > 240 {
		message = message[:240]
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE upstream_sources
		SET last_error = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, message, id)
	return err
}

func (s *Store) uniqueAutoPrefix(ctx context.Context, name, sourceURL string, excludeID int64) (string, error) {
	for sequence := 1; sequence < 1000; sequence++ {
		prefix := naming.AutoPrefix(name, sourceURL, fmt.Sprintf("Source-%d", sequence), sequence)
		var count int
		err := s.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM upstream_sources
			WHERE display_prefix = ?
			  AND id != ?
		`, prefix, excludeID).Scan(&count)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return prefix, nil
		}
	}
	return "", fmt.Errorf("unable to allocate unique prefix for source %q", name)
}

func sourceExists(err error) bool {
	return err == nil || !errorsIsNoRows(err)
}

func errorsIsNoRows(err error) bool {
	return err == sql.ErrNoRows
}
