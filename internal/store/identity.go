package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/security"
)

type CreateTeamInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateUserInput struct {
	TeamID *int64 `json:"team_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Remark string `json:"remark"`
}

type CreateTokenInput struct {
	UserID     int64  `json:"user_id"`
	Name       string `json:"name"`
	ExpireDays int    `json:"expire_days"`
	QuotaBytes int64  `json:"quota_bytes"`
}

type CreateTokenResult struct {
	Token         Token              `json:"token"`
	PlainToken    string             `json:"plain_token"`
	Subscription  string             `json:"subscription"`
	Subscriptions TokenSubscriptions `json:"subscriptions"`
	Account       GatewayAccount     `json:"gateway_account"`
}

type TokenSubscriptions struct {
	Default string `json:"default"`
	Clash   string `json:"clash"`
	SingBox string `json:"sing_box"`
}

func (s *Store) CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO teams(name, description)
		VALUES (?, ?)
	`, input.Name, input.Description)
	if err != nil {
		return Team{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Team{}, err
	}
	return s.GetTeam(ctx, id)
}

func (s *Store) GetTeam(ctx context.Context, id int64) (Team, error) {
	var team Team
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, status, created_at, updated_at
		FROM teams
		WHERE id = ?
	`, id).Scan(&team.ID, &team.Name, &team.Description, &team.Status, &team.CreatedAt, &team.UpdatedAt)
	return team, err
}

func (s *Store) ListTeams(ctx context.Context) ([]Team, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, status, created_at, updated_at
		FROM teams
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.Name, &team.Description, &team.Status, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (s *Store) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO users(team_id, name, email, remark)
		VALUES (?, ?, ?, ?)
	`, input.TeamID, input.Name, input.Email, input.Remark)
	if err != nil {
		return User{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return User{}, err
	}
	return s.GetUser(ctx, id)
}

func (s *Store) GetUser(ctx context.Context, id int64) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, team_id, name, email, remark, status, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &user.TeamID, &user.Name, &user.Email, &user.Remark, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, team_id, name, email, remark, status, created_at, updated_at
		FROM users
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.TeamID, &user.Name, &user.Email, &user.Remark, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) CreateToken(ctx context.Context, secret, publicBaseURL string, input CreateTokenInput) (CreateTokenResult, error) {
	plainToken, err := security.NewToken()
	if err != nil {
		return CreateTokenResult{}, err
	}
	uuid, err := security.NewUUID()
	if err != nil {
		return CreateTokenResult{}, err
	}

	var expireAt any
	if input.ExpireDays > 0 {
		expireAt = time.Now().UTC().AddDate(0, 0, input.ExpireDays).Format(time.RFC3339)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateTokenResult{}, err
	}
	defer rollback(tx)

	result, err := tx.ExecContext(ctx, `
		INSERT INTO tokens(user_id, token_hash, token_prefix, name, expire_at, quota_bytes)
		VALUES (?, ?, ?, ?, ?, ?)
	`, input.UserID, security.TokenHash(secret, plainToken), security.TokenPrefix(plainToken), input.Name, expireAt, input.QuotaBytes)
	if err != nil {
		return CreateTokenResult{}, err
	}
	tokenID, err := result.LastInsertId()
	if err != nil {
		return CreateTokenResult{}, err
	}

	authUser := fmt.Sprintf("fg_u_%d_t_%d", input.UserID, tokenID)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO gateway_accounts(token_id, protocol, auth_user, uuid)
		VALUES (?, 'vless', ?, ?)
	`, tokenID, authUser, uuid); err != nil {
		return CreateTokenResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return CreateTokenResult{}, err
	}

	token, err := s.GetToken(ctx, tokenID)
	if err != nil {
		return CreateTokenResult{}, err
	}
	account, err := s.GetGatewayAccountByToken(ctx, tokenID)
	if err != nil {
		return CreateTokenResult{}, err
	}

	subscriptions := tokenSubscriptionURLs(publicBaseURL, plainToken)
	return CreateTokenResult{
		Token:         token,
		PlainToken:    plainToken,
		Subscription:  subscriptions.Default,
		Subscriptions: subscriptions,
		Account:       account,
	}, nil
}

func tokenSubscriptionURLs(publicBaseURL, plainToken string) TokenSubscriptions {
	base := strings.TrimRight(publicBaseURL, "/") + "/sub/" + plainToken
	return TokenSubscriptions{
		Default: base,
		Clash:   base + "?target=clash",
		SingBox: base + "?target=sing-box",
	}
}

func (s *Store) GetToken(ctx context.Context, id int64) (Token, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_prefix, name, status, expire_at, quota_bytes,
		       used_upload_bytes, used_download_bytes, last_used_at, created_at, updated_at, revoked_at
		FROM tokens
		WHERE id = ?
	`, id)
	return scanToken(row)
}

func (s *Store) TokenByHash(ctx context.Context, tokenHash string) (TokenWithAccount, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT t.id, t.user_id, t.token_prefix, t.name, t.status, t.expire_at, t.quota_bytes,
		       t.used_upload_bytes, t.used_download_bytes, t.last_used_at, t.created_at, t.updated_at, t.revoked_at,
		       u.team_id,
		       g.id, g.token_id, g.protocol, g.auth_user, g.uuid, g.password, g.status, g.created_at, g.updated_at
		FROM tokens t
		JOIN users u ON u.id = t.user_id
		JOIN gateway_accounts g ON g.token_id = t.id
		WHERE t.token_hash = ?
		  AND u.status = 'active'
	`, tokenHash)
	return scanTokenWithAccount(row)
}

func (s *Store) ListTokens(ctx context.Context) ([]TokenWithAccount, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.user_id, t.token_prefix, t.name, t.status, t.expire_at, t.quota_bytes,
		       t.used_upload_bytes, t.used_download_bytes, t.last_used_at, t.created_at, t.updated_at, t.revoked_at,
		       u.team_id,
		       g.id, g.token_id, g.protocol, g.auth_user, g.uuid, g.password, g.status, g.created_at, g.updated_at
		FROM tokens t
		JOIN users u ON u.id = t.user_id
		JOIN gateway_accounts g ON g.token_id = t.id
		ORDER BY t.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tokens []TokenWithAccount
	for rows.Next() {
		token, err := scanTokenWithAccount(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (s *Store) RevokeToken(ctx context.Context, id int64) (Token, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Token{}, err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET status = 'revoked', revoked_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id); err != nil {
		return Token{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gateway_accounts
		SET status = 'revoked', updated_at = CURRENT_TIMESTAMP
		WHERE token_id = ?
	`, id); err != nil {
		return Token{}, err
	}
	if err := tx.Commit(); err != nil {
		return Token{}, err
	}
	return s.GetToken(ctx, id)
}

func (s *Store) RestoreToken(ctx context.Context, id int64) (Token, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Token{}, err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET status = 'active', revoked_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, id); err != nil {
		return Token{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gateway_accounts
		SET status = 'active', updated_at = CURRENT_TIMESTAMP
		WHERE token_id = ?
	`, id); err != nil {
		return Token{}, err
	}
	if err := tx.Commit(); err != nil {
		return Token{}, err
	}
	return s.GetToken(ctx, id)
}

func (s *Store) ExtendToken(ctx context.Context, id int64, days int) (Token, error) {
	if days <= 0 {
		return Token{}, fmt.Errorf("extend_days must be greater than 0")
	}
	token, err := s.GetToken(ctx, id)
	if err != nil {
		return Token{}, err
	}
	base := time.Now().UTC()
	if token.ExpireAt != nil && token.ExpireAt.After(base) {
		base = token.ExpireAt.UTC()
	}
	expireAt := base.AddDate(0, 0, days).Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE tokens
		SET expire_at = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, expireAt, id); err != nil {
		return Token{}, err
	}
	return s.GetToken(ctx, id)
}

func (s *Store) AddTokenQuota(ctx context.Context, id int64, quotaBytes int64) (Token, error) {
	if quotaBytes <= 0 {
		return Token{}, fmt.Errorf("quota_bytes must be greater than 0")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Token{}, err
	}
	defer rollback(tx)

	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET quota_bytes = quota_bytes + ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, quotaBytes, id); err != nil {
		return Token{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tokens
		SET status = 'active', updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		  AND status = 'over_quota'
		  AND quota_bytes > 0
		  AND used_upload_bytes + used_download_bytes < quota_bytes
	`, id); err != nil {
		return Token{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE gateway_accounts
		SET status = 'active', updated_at = CURRENT_TIMESTAMP
		WHERE token_id = ?
		  AND status = 'over_quota'
		  AND EXISTS (
		    SELECT 1 FROM tokens
		    WHERE tokens.id = gateway_accounts.token_id
		      AND tokens.status = 'active'
		  )
	`, id); err != nil {
		return Token{}, err
	}
	if err := tx.Commit(); err != nil {
		return Token{}, err
	}
	return s.GetToken(ctx, id)
}

func (s *Store) TouchTokenUsed(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "UPDATE tokens SET last_used_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func (s *Store) GetGatewayAccountByToken(ctx context.Context, tokenID int64) (GatewayAccount, error) {
	var account GatewayAccount
	err := s.db.QueryRowContext(ctx, `
		SELECT id, token_id, protocol, auth_user, uuid, password, status, created_at, updated_at
		FROM gateway_accounts
		WHERE token_id = ?
	`, tokenID).Scan(&account.ID, &account.TokenID, &account.Protocol, &account.AuthUser, &account.UUID, &account.Password, &account.Status, &account.CreatedAt, &account.UpdatedAt)
	return account, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanToken(scanner scanner) (Token, error) {
	var token Token
	var expireAt sql.NullString
	err := scanner.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenPrefix,
		&token.Name,
		&token.Status,
		&expireAt,
		&token.QuotaBytes,
		&token.UsedUploadBytes,
		&token.UsedDownloadBytes,
		&token.LastUsedAt,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.RevokedAt,
	)
	if err != nil {
		return Token{}, err
	}
	if expireAt.Valid && expireAt.String != "" {
		parsed, err := time.Parse(time.RFC3339, expireAt.String)
		if err == nil {
			token.ExpireAt = &parsed
		}
	}
	return token, nil
}

func scanTokenWithAccount(scanner scanner) (TokenWithAccount, error) {
	var item TokenWithAccount
	var expireAt sql.NullString
	var userTeamID sql.NullInt64
	err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.TokenPrefix,
		&item.Name,
		&item.Status,
		&expireAt,
		&item.QuotaBytes,
		&item.UsedUploadBytes,
		&item.UsedDownloadBytes,
		&item.LastUsedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.RevokedAt,
		&userTeamID,
		&item.GatewayAccount.ID,
		&item.GatewayAccount.TokenID,
		&item.GatewayAccount.Protocol,
		&item.GatewayAccount.AuthUser,
		&item.GatewayAccount.UUID,
		&item.GatewayAccount.Password,
		&item.GatewayAccount.Status,
		&item.GatewayAccount.CreatedAt,
		&item.GatewayAccount.UpdatedAt,
	)
	if err != nil {
		return TokenWithAccount{}, err
	}
	if expireAt.Valid && expireAt.String != "" {
		parsed, err := time.Parse(time.RFC3339, expireAt.String)
		if err == nil {
			item.ExpireAt = &parsed
		}
	}
	if userTeamID.Valid {
		item.UserTeamID = &userTeamID.Int64
	}
	return item, nil
}
