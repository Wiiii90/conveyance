package sqlite

import "database/sql"

const schema = `
CREATE TABLE IF NOT EXISTS current_objects (
    trust_domain_ref TEXT NOT NULL,
    channel_ref TEXT NOT NULL,
    envelope_format_version INTEGER NOT NULL CHECK (envelope_format_version <> 0),
    channel_epoch INTEGER NOT NULL CHECK (channel_epoch <> 0),
    revision INTEGER NOT NULL CHECK (revision <> 0),
    envelope_ref TEXT NOT NULL,
    previous_envelope_ref TEXT,
    protected_payload BLOB NOT NULL,
    PRIMARY KEY (trust_domain_ref, channel_ref)
);`

func initializeSchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}
