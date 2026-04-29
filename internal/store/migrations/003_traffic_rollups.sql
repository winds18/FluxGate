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

CREATE TABLE IF NOT EXISTS traffic_user_hourly (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hour TEXT NOT NULL,
  user_id INTEGER NOT NULL,
  token_id INTEGER NOT NULL,
  gateway_account_id INTEGER NOT NULL,
  upload_bytes INTEGER NOT NULL DEFAULT 0,
  download_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(token_id) REFERENCES tokens(id),
  FOREIGN KEY(gateway_account_id) REFERENCES gateway_accounts(id),
  UNIQUE(hour, token_id)
);

CREATE TABLE IF NOT EXISTS traffic_user_daily (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  day TEXT NOT NULL,
  user_id INTEGER NOT NULL,
  token_id INTEGER NOT NULL,
  gateway_account_id INTEGER NOT NULL,
  upload_bytes INTEGER NOT NULL DEFAULT 0,
  download_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(token_id) REFERENCES tokens(id),
  FOREIGN KEY(gateway_account_id) REFERENCES gateway_accounts(id),
  UNIQUE(day, token_id)
);

CREATE TABLE IF NOT EXISTS traffic_outbound_daily (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  day TEXT NOT NULL,
  upstream_node_id INTEGER,
  outbound_tag TEXT NOT NULL,
  upload_bytes INTEGER NOT NULL DEFAULT 0,
  download_bytes INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(upstream_node_id) REFERENCES upstream_nodes(id),
  UNIQUE(day, outbound_tag)
);

CREATE INDEX IF NOT EXISTS idx_traffic_samples_metric_time ON traffic_samples(metric_type, metric_name, sampled_at);
CREATE INDEX IF NOT EXISTS idx_traffic_user_hourly_token_hour ON traffic_user_hourly(token_id, hour);
CREATE INDEX IF NOT EXISTS idx_traffic_user_daily_token_day ON traffic_user_daily(token_id, day);
CREATE INDEX IF NOT EXISTS idx_traffic_outbound_daily_tag_day ON traffic_outbound_daily(outbound_tag, day);
