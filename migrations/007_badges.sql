-- V7: 成就徽章体系
-- 用户已获得的徽章；唯一索引保证同一用户同一徽章只有一枚（重复达成不发第二枚）
CREATE TABLE IF NOT EXISTS user_badges (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  badge_code VARCHAR(50) NOT NULL,
  award_event VARCHAR(30) NOT NULL DEFAULT 'sweep',
  metric_snapshot TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_user_badges_user ON user_badges(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_user_badge ON user_badges(user_id, badge_code);

-- 徽章判定/发放流水：granted / duplicate / not_met 全部可追踪
CREATE TABLE IF NOT EXISTS badge_award_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  badge_code VARCHAR(50) NOT NULL,
  event VARCHAR(30) NOT NULL,
  result VARCHAR(20) NOT NULL,
  metric_snapshot TEXT,
  remark VARCHAR(255),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_badge_logs_user ON badge_award_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_badge_logs_badge ON badge_award_logs(badge_code);
CREATE INDEX IF NOT EXISTS idx_badge_logs_result ON badge_award_logs(result);
