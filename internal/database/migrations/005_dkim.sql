CREATE TABLE IF NOT EXISTS dkim_keys (
    domain TEXT NOT NULL,
    selector TEXT NOT NULL,
    public_key_pem TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (domain, selector)
);

CREATE INDEX IF NOT EXISTS idx_dkim_keys_domain ON dkim_keys(domain);
