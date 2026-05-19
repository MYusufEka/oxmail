CREATE TABLE IF NOT EXISTS aliases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_address TEXT NOT NULL,
    destination_address TEXT NOT NULL,
    domain_id INTEGER NOT NULL,
    active BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_aliases_source ON aliases(source_address);
CREATE INDEX IF NOT EXISTS idx_aliases_domain ON aliases(domain_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_aliases_source_dest ON aliases(source_address, destination_address);
