-- V7: admin users and asynchronous audit log exports
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS audit_exports (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  requester_id INTEGER NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'queued',
  user_id INTEGER NOT NULL DEFAULT 0,
  operator VARCHAR(100) NOT NULL DEFAULT '',
  action VARCHAR(50) NOT NULL DEFAULT '',
  target_type VARCHAR(30) NOT NULL DEFAULT '',
  from_time DATETIME,
  to_time DATETIME,
  file_path VARCHAR(500) NOT NULL DEFAULT '',
  file_name VARCHAR(200) NOT NULL DEFAULT '',
  row_count INTEGER NOT NULL DEFAULT 0,
  error_category VARCHAR(20) NOT NULL DEFAULT '',
  error_message VARCHAR(500) NOT NULL DEFAULT '',
  started_at DATETIME,
  finished_at DATETIME,
  expires_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_audit_exports_requester ON audit_exports(requester_id);
CREATE INDEX IF NOT EXISTS idx_audit_exports_status ON audit_exports(status);
CREATE INDEX IF NOT EXISTS idx_audit_exports_expires ON audit_exports(expires_at);
