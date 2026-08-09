package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Wiiii90/conveyance/internal/currentobject"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
}

// Open opens path as a single-node SQLite Current Object store and initializes
// the v0.1.0 schema. The caller owns the returned repository's lifecycle.
func Open(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w: %w", currentobject.ErrUnavailable, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure sqlite database: %w: %w", currentobject.ErrUnavailable, err)
	}
	if err := initializeSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize sqlite schema: %w: %w", currentobject.ErrUnavailable, err)
	}
	return &Repository{db: db}, nil
}

func (repository *Repository) Close() error {
	if repository == nil || repository.db == nil {
		return nil
	}
	return repository.db.Close()
}

func (repository *Repository) GetCurrent(ctx context.Context, trustDomainRef currentobject.TrustDomainRef, channelRef currentobject.ChannelRef) (currentobject.Envelope, error) {
	const query = `SELECT envelope_format_version, channel_epoch, revision,
        envelope_ref, previous_envelope_ref, protected_payload
        FROM current_objects WHERE trust_domain_ref = ? AND channel_ref = ?`

	var formatVersion, epoch, revision int64
	var envelopeRef, previousEnvelopeRef sql.NullString
	var payload []byte
	err := repository.db.QueryRowContext(ctx, query, trustDomainRef.String(), channelRef.String()).Scan(
		&formatVersion, &epoch, &revision, &envelopeRef, &previousEnvelopeRef, &payload,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return currentobject.Envelope{}, currentobject.ErrCurrentObjectNotFound
	}
	if err != nil {
		return currentobject.Envelope{}, unavailable("get current object", err)
	}
	previous, err := parseOptionalEnvelopeRef(previousEnvelopeRef)
	if err != nil {
		return currentobject.Envelope{}, unavailable("decode current object", err)
	}
	ref, err := currentobject.ParseEnvelopeRef(envelopeRef.String)
	if err != nil {
		return currentobject.Envelope{}, unavailable("decode current object", err)
	}
	return currentobject.NewEnvelope(decodeUint64(formatVersion), decodeUint64(epoch), decodeUint64(revision), ref, previous, payload), nil
}

func (repository *Repository) CompareAndSwapCurrent(ctx context.Context, trustDomainRef currentobject.TrustDomainRef, channelRef currentobject.ChannelRef, expected *currentobject.CurrentState, next currentobject.Envelope) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return unavailable("begin current object transaction", err)
	}
	defer func() { _ = tx.Rollback() }()

	var result sql.Result
	if expected == nil {
		result, err = tx.ExecContext(ctx, `INSERT INTO current_objects
            (trust_domain_ref, channel_ref, envelope_format_version, channel_epoch,
             revision, envelope_ref, previous_envelope_ref, protected_payload)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			trustDomainRef.String(), channelRef.String(), encodeUint64(next.EnvelopeFormatVersion),
			encodeUint64(next.ChannelEpoch), encodeUint64(next.Revision), next.EnvelopeRef.String(),
			optionalEnvelopeRef(next.PreviousEnvelopeRef), append([]byte(nil), next.ProtectedPayload...))
		if err != nil {
			if isConstraintError(err) {
				return currentobject.ErrConflict
			}
			return unavailable("insert current object", err)
		}
	} else {
		result, err = tx.ExecContext(ctx, `UPDATE current_objects SET
            envelope_format_version = ?, channel_epoch = ?, revision = ?,
            envelope_ref = ?, previous_envelope_ref = ?, protected_payload = ?
            WHERE trust_domain_ref = ? AND channel_ref = ?
              AND channel_epoch = ? AND revision = ? AND envelope_ref = ?`,
			encodeUint64(next.EnvelopeFormatVersion), encodeUint64(next.ChannelEpoch), encodeUint64(next.Revision),
			next.EnvelopeRef.String(), optionalEnvelopeRef(next.PreviousEnvelopeRef),
			append([]byte(nil), next.ProtectedPayload...), trustDomainRef.String(), channelRef.String(),
			encodeUint64(expected.ChannelEpoch), encodeUint64(expected.Revision), expected.EnvelopeRef.String())
		if err != nil {
			return unavailable("update current object", err)
		}
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return unavailable("inspect current object write", err)
	}
	if rows != 1 {
		return currentobject.ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return unavailable("commit current object write", err)
	}
	return nil
}

// SQLite INTEGER is signed int64. Encoding by bit pattern preserves every
// uint64 value for storage and equality comparison, including MaxUint64.
func encodeUint64(value uint64) int64 { return int64(value) }

func decodeUint64(value int64) uint64 { return uint64(value) }

func parseOptionalEnvelopeRef(value sql.NullString) (*currentobject.EnvelopeRef, error) {
	if !value.Valid {
		return nil, nil
	}
	ref, err := currentobject.ParseEnvelopeRef(value.String)
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func optionalEnvelopeRef(value *currentobject.EnvelopeRef) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func unavailable(operation string, err error) error {
	return fmt.Errorf("%s: %w: %w", operation, currentobject.ErrUnavailable, err)
}

func isConstraintError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}
