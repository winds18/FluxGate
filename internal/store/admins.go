package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/winds18/FluxGate/internal/security"
)

type BootstrapAdminResult struct {
	Admin   Admin
	Created bool
	Skipped bool
}

func (s *Store) BootstrapAdmin(ctx context.Context, username, password string) (BootstrapAdminResult, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return BootstrapAdminResult{Skipped: true}, nil
	}

	var count int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM admins").Scan(&count); err != nil {
		return BootstrapAdminResult{}, err
	}
	if count > 0 {
		return BootstrapAdminResult{Skipped: true}, nil
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO admins(username, password_hash)
		VALUES (?, ?)
	`, username, hash)
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	admin, err := s.GetAdmin(ctx, id)
	if err != nil {
		return BootstrapAdminResult{}, err
	}
	return BootstrapAdminResult{Admin: admin, Created: true}, nil
}

func (s *Store) GetAdmin(ctx context.Context, id int64) (Admin, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, status, last_login_at, created_at, updated_at
		FROM admins
		WHERE id = ?
	`, id)
	return scanAdmin(row)
}

func (s *Store) AdminByUsername(ctx context.Context, username string) (AdminWithPassword, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, status, last_login_at, created_at, updated_at
		FROM admins
		WHERE username = ?
	`, strings.TrimSpace(username))
	var admin AdminWithPassword
	err := row.Scan(
		&admin.ID,
		&admin.Username,
		&admin.PasswordHash,
		&admin.Status,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	return admin, err
}

func (s *Store) AuthenticateAdmin(ctx context.Context, username, password string) (Admin, error) {
	admin, err := s.AdminByUsername(ctx, username)
	if err != nil {
		return Admin{}, err
	}
	if admin.Status != "active" {
		return Admin{}, fmt.Errorf("admin is not active")
	}
	if !security.VerifyPassword(password, admin.PasswordHash) {
		return Admin{}, sql.ErrNoRows
	}
	if _, err := s.db.ExecContext(ctx, `
		UPDATE admins
		SET last_login_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, admin.ID); err != nil {
		return Admin{}, err
	}
	return s.GetAdmin(ctx, admin.ID)
}

func scanAdmin(row interface {
	Scan(dest ...any) error
}) (Admin, error) {
	var admin Admin
	err := row.Scan(
		&admin.ID,
		&admin.Username,
		&admin.Status,
		&admin.LastLoginAt,
		&admin.CreatedAt,
		&admin.UpdatedAt,
	)
	return admin, err
}
