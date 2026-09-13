-- V3: tutorials
CREATE TABLE IF NOT EXISTS tutorials (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(200),
  summary TEXT,
  cover_before VARCHAR(255) NOT NULL,
  cover_after VARCHAR(255) NOT NULL,
  difficulty VARCHAR(20) NOT NULL DEFAULT 'medium',
  estimated_hours REAL NOT NULL DEFAULT 1,
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  version INTEGER NOT NULL DEFAULT 1,
  view_count INTEGER NOT NULL DEFAULT 0,
  favorite_count INTEGER NOT NULL DEFAULT 0,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  comment_count INTEGER NOT NULL DEFAULT 0,
  project_count INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_tutorials_user ON tutorials(user_id);
CREATE INDEX IF NOT EXISTS idx_tutorials_cat ON tutorials(category_id);
CREATE INDEX IF NOT EXISTS idx_tutorials_status ON tutorials(status);
CREATE INDEX IF NOT EXISTS idx_tutorials_views ON tutorials(view_count DESC);
CREATE INDEX IF NOT EXISTS idx_tutorials_fav ON tutorials(favorite_count DESC);
CREATE TABLE IF NOT EXISTS tutorial_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  tutorial_id INTEGER NOT NULL,
  version INTEGER NOT NULL,
  title VARCHAR(200),
  summary TEXT,
  content_dump TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tv_tid ON tutorial_versions(tutorial_id);
CREATE TABLE IF NOT EXISTS tutorial_tags (
  tutorial_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (tutorial_id, tag_id)
);
