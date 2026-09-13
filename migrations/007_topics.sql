-- V7: topics, topic_items（专题模块）
CREATE TABLE IF NOT EXISTS topics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title VARCHAR(100) NOT NULL,
  summary VARCHAR(500),
  cover VARCHAR(255),
  status INTEGER NOT NULL DEFAULT 1,
  item_count INTEGER NOT NULL DEFAULT 0,
  created_by INTEGER,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_topics_status ON topics(status);
CREATE INDEX IF NOT EXISTS idx_topics_created_by ON topics(created_by);

CREATE TABLE IF NOT EXISTS topic_items (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  topic_id INTEGER NOT NULL,
  item_type VARCHAR(20) NOT NULL,
  item_id INTEGER NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_topic_items_topic ON topic_items(topic_id);
-- 同一条目在同一专题内只保留一次
CREATE UNIQUE INDEX IF NOT EXISTS idx_topic_item_unique ON topic_items(topic_id, item_type, item_id);
