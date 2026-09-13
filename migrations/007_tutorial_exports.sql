-- V7: tutorial_exports (A4 打印快照记录)
CREATE TABLE IF NOT EXISTS tutorial_exports (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tutorial_id INTEGER NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  title VARCHAR(200),
  file_name VARCHAR(255) NOT NULL,
  file_url VARCHAR(500) NOT NULL,
  file_size INTEGER NOT NULL DEFAULT 0,
  created_by INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tutorial_exports_tid ON tutorial_exports(tutorial_id);
CREATE INDEX IF NOT EXISTS idx_tutorial_exports_creator ON tutorial_exports(created_by);
