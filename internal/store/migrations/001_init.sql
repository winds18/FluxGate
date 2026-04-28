CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admins (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  last_login_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teams (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  team_id INTEGER,
  name TEXT NOT NULL,
  email TEXT NOT NULL DEFAULT '',
  remark TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(team_id) REFERENCES teams(id)
);

CREATE TABLE IF NOT EXISTS tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  token_prefix TEXT NOT NULL,
  name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  expire_at TEXT,
  quota_bytes INTEGER NOT NULL DEFAULT 0,
  used_upload_bytes INTEGER NOT NULL DEFAULT 0,
  used_download_bytes INTEGER NOT NULL DEFAULT 0,
  last_used_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  revoked_at TEXT,
  FOREIGN KEY(user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS gateway_accounts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token_id INTEGER NOT NULL UNIQUE,
  protocol TEXT NOT NULL DEFAULT 'vless',
  auth_user TEXT NOT NULL UNIQUE,
  uuid TEXT NOT NULL,
  password TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  rotated_at TEXT,
  FOREIGN KEY(token_id) REFERENCES tokens(id)
);

CREATE TABLE IF NOT EXISTS upstream_sources (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'manual',
  url TEXT NOT NULL DEFAULT '',
  raw_content TEXT NOT NULL DEFAULT '',
  prefix_mode TEXT NOT NULL DEFAULT 'auto',
  display_prefix TEXT NOT NULL,
  default_tags TEXT NOT NULL DEFAULT '[]',
  refresh_interval_minutes INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  last_sync_at TEXT,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS upstream_nodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_id INTEGER NOT NULL,
  raw_name TEXT NOT NULL,
  display_name TEXT NOT NULL,
  name_mode TEXT NOT NULL DEFAULT 'auto',
  uri TEXT NOT NULL,
  uri_hash TEXT NOT NULL,
  protocol TEXT NOT NULL DEFAULT '',
  server TEXT NOT NULL DEFAULT '',
  server_port INTEGER NOT NULL DEFAULT 0,
  region TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  last_seen_at TEXT,
  last_checked_at TEXT,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(source_id) REFERENCES upstream_sources(id),
  UNIQUE(source_id, uri_hash)
);

CREATE TABLE IF NOT EXISTS tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  color TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS upstream_node_tags (
  node_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  PRIMARY KEY(node_id, tag_id),
  FOREIGN KEY(node_id) REFERENCES upstream_nodes(id),
  FOREIGN KEY(tag_id) REFERENCES tags(id)
);

CREATE TABLE IF NOT EXISTS virtual_nodes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  listen_protocol TEXT NOT NULL DEFAULT 'vless',
  listen_port INTEGER NOT NULL,
  tag_selector TEXT NOT NULL DEFAULT '{}',
  strategy TEXT NOT NULL DEFAULT 'selector',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscription_access_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token_id INTEGER,
  ip TEXT NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  target TEXT NOT NULL DEFAULT '',
  status_code INTEGER NOT NULL,
  message TEXT NOT NULL DEFAULT '',
  response_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(token_id) REFERENCES tokens(id)
);

CREATE TABLE IF NOT EXISTS config_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  version TEXT NOT NULL,
  config_hash TEXT NOT NULL,
  config_path TEXT NOT NULL,
  status TEXT NOT NULL,
  check_output TEXT NOT NULL DEFAULT '',
  published_at TEXT,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS traffic_samples (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  sampled_at TEXT NOT NULL,
  metric_type TEXT NOT NULL,
  metric_name TEXT NOT NULL,
  upload_bytes_delta INTEGER NOT NULL DEFAULT 0,
  download_bytes_delta INTEGER NOT NULL DEFAULT 0,
  raw_value_upload INTEGER NOT NULL DEFAULT 0,
  raw_value_download INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tokens_hash ON tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_status_expire ON tokens(status, expire_at);
CREATE INDEX IF NOT EXISTS idx_gateway_accounts_auth_user ON gateway_accounts(auth_user);
CREATE INDEX IF NOT EXISTS idx_upstream_nodes_uri_hash ON upstream_nodes(uri_hash);
CREATE INDEX IF NOT EXISTS idx_upstream_nodes_source ON upstream_nodes(source_id);
CREATE INDEX IF NOT EXISTS idx_subscription_access_logs_token_time ON subscription_access_logs(token_id, created_at);
