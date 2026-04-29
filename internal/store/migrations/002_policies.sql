CREATE TABLE IF NOT EXISTS policies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  scope_type TEXT NOT NULL,
  scope_id INTEGER,
  include_tags TEXT NOT NULL DEFAULT '[]',
  exclude_tags TEXT NOT NULL DEFAULT '[]',
  allowed_virtual_nodes TEXT NOT NULL DEFAULT '[]',
  max_nodes INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CHECK(scope_type IN ('team', 'user', 'token')),
  CHECK(max_nodes >= 0)
);

CREATE INDEX IF NOT EXISTS idx_policies_scope ON policies(scope_type, scope_id);
CREATE INDEX IF NOT EXISTS idx_policies_status ON policies(status);
