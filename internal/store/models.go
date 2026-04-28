package store

import "time"

type Overview struct {
	Teams        int64  `json:"teams"`
	Users        int64  `json:"users"`
	Tokens       int64  `json:"tokens"`
	Sources      int64  `json:"sources"`
	Nodes        int64  `json:"nodes"`
	VirtualNodes int64  `json:"virtual_nodes"`
	Version      string `json:"version"`
}

type Admin struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Status      string  `json:"status"`
	LastLoginAt *string `json:"last_login_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type AdminWithPassword struct {
	Admin
	PasswordHash string `json:"-"`
}

type Source struct {
	ID                     int64   `json:"id"`
	Name                   string  `json:"name"`
	Type                   string  `json:"type"`
	URL                    string  `json:"url"`
	RawContent             string  `json:"raw_content,omitempty"`
	PrefixMode             string  `json:"prefix_mode"`
	DisplayPrefix          string  `json:"display_prefix"`
	DefaultTags            string  `json:"default_tags"`
	RefreshIntervalMinutes int64   `json:"refresh_interval_minutes"`
	Status                 string  `json:"status"`
	LastSyncAt             *string `json:"last_sync_at,omitempty"`
	LastError              string  `json:"last_error"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

type Node struct {
	ID            int64   `json:"id"`
	SourceID      int64   `json:"source_id"`
	SourceName    string  `json:"source_name,omitempty"`
	RawName       string  `json:"raw_name"`
	DisplayName   string  `json:"display_name"`
	NameMode      string  `json:"name_mode"`
	URI           string  `json:"uri,omitempty"`
	URIHash       string  `json:"uri_hash"`
	Protocol      string  `json:"protocol"`
	Server        string  `json:"server"`
	ServerPort    int     `json:"server_port"`
	Region        string  `json:"region"`
	Status        string  `json:"status"`
	LastSeenAt    *string `json:"last_seen_at,omitempty"`
	LastCheckedAt *string `json:"last_checked_at,omitempty"`
	LastError     string  `json:"last_error"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type Team struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type User struct {
	ID        int64  `json:"id"`
	TeamID    *int64 `json:"team_id,omitempty"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Remark    string `json:"remark"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Token struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	TokenPrefix       string     `json:"token_prefix"`
	Name              string     `json:"name"`
	Status            string     `json:"status"`
	ExpireAt          *time.Time `json:"expire_at,omitempty"`
	QuotaBytes        int64      `json:"quota_bytes"`
	UsedUploadBytes   int64      `json:"used_upload_bytes"`
	UsedDownloadBytes int64      `json:"used_download_bytes"`
	LastUsedAt        *string    `json:"last_used_at,omitempty"`
	CreatedAt         string     `json:"created_at"`
	UpdatedAt         string     `json:"updated_at"`
	RevokedAt         *string    `json:"revoked_at,omitempty"`
}

type GatewayAccount struct {
	ID        int64  `json:"id"`
	TokenID   int64  `json:"token_id"`
	Protocol  string `json:"protocol"`
	AuthUser  string `json:"auth_user"`
	UUID      string `json:"uuid"`
	Password  string `json:"password,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TokenWithAccount struct {
	Token
	GatewayAccount GatewayAccount `json:"gateway_account"`
}

type VirtualNode struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	ListenProtocol string `json:"listen_protocol"`
	ListenPort     int    `json:"listen_port"`
	TagSelector    string `json:"tag_selector"`
	Strategy       string `json:"strategy"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type ImportResult struct {
	Imported    int `json:"imported"`
	Updated     int `json:"updated"`
	Skipped     int `json:"skipped"`
	Inactivated int `json:"inactivated"`
}
