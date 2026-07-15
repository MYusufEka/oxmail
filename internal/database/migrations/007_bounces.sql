CREATE TABLE IF NOT EXISTS bounces (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recipient TEXT NOT NULL,
  sender TEXT NOT NULL,
  subject TEXT DEFAULT '',
  bounce_type TEXT NOT NULL CHECK(bounce_type IN ('hard','soft')),
  error_message TEXT NOT NULL,
  bounced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_bounces_recipient ON bounces(recipient);
CREATE INDEX IF NOT EXISTS idx_bounces_bounced_at ON bounces(bounced_at);
