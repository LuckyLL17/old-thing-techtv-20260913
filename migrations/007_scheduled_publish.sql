-- V7: scheduled publish
ALTER TABLE tutorials ADD COLUMN scheduled_at DATETIME;
ALTER TABLE tutorials ADD COLUMN published_at DATETIME;
CREATE INDEX IF NOT EXISTS idx_tutorials_sched ON tutorials(status, scheduled_at);
