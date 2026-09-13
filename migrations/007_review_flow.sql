-- V7: tutorial review flow
-- 教程审核流：用户角色 + 教程审核字段
ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
ALTER TABLE tutorials ADD COLUMN review_note VARCHAR(500);
ALTER TABLE tutorials ADD COLUMN reviewed_at DATETIME;
