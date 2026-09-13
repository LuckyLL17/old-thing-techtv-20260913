-- V5: projects, comments, favorites, attempts
CREATE TABLE IF NOT EXISTS projects (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  tutorial_id INTEGER NOT NULL,
  title VARCHAR(200),
  description TEXT,
  images TEXT,
  custom_notes TEXT,
  rating INTEGER NOT NULL DEFAULT 0,
  like_count INTEGER NOT NULL DEFAULT 0,
  comment_count INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_projects_uid ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_tid ON projects(tutorial_id);
CREATE INDEX IF NOT EXISTS idx_projects_rating ON projects(rating DESC);
CREATE TABLE IF NOT EXISTS comments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  target_type VARCHAR(20) NOT NULL,
  target_id INTEGER NOT NULL,
  parent_id INTEGER NOT NULL DEFAULT 0,
  content TEXT NOT NULL,
  like_count INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);
CREATE TABLE IF NOT EXISTS favorites (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  target_type VARCHAR(20) NOT NULL,
  target_id INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_fav_user ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_fav_target ON favorites(target_type, target_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fav_unique ON favorites(user_id, target_type, target_id);
CREATE TABLE IF NOT EXISTS attempts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  tutorial_id INTEGER NOT NULL,
  status INTEGER NOT NULL DEFAULT 1,
  completed INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_attempts_user ON attempts(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_attempts_unique ON attempts(user_id, tutorial_id);
