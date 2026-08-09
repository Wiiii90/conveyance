# 02 -- Scenarios

## S1 -- Publish Vocation Current Object

Given Vocation exposes a valid `Published Opportunity Overview 1.0`,
when WGT Windows validates and protects that payload, then it can
publish one opaque Current Object to an authorized Conveyance Channel.

Conveyance does not parse or validate the Vocation contract.

## S2 -- Read while source PC is offline

Given a Current Object was successfully published, and the Windows PC is
offline, when an authorized iPhone installation requests the Channel,
then Conveyance returns the current opaque Envelope. The iPhone
verifies/decrypts it, validates the Vocation contract, and may cache it
locally.

## S3 -- Concurrent replacement

Given Envelope A is current, when two publishers independently attempt
replacement based on A, then only one replacement may succeed. The other
receives a generic conflict and must re-read current state.

## S4 -- Replay to an established installation

Given a client has accepted `(epoch=4, revision=9)`, when a server later
presents `(epoch=4, revision=8)`, then the client rejects the rollback.

## S5 -- Lost installation

Given an enrolled iPhone is lost, when it is revoked, then Conveyance
rejects its future authenticated operations and affected Channels move
to a fresh epoch/key before confidential future content is published.

Previously available old-epoch plaintext cannot be cryptographically
revoked.

## S6 -- Enroll a new installation

A new installation creates fresh local identity/key material. An already
trusted installation authorizes enrollment. The server may mediate the
exchange but cannot authorize trust by itself. Current Channel keys are
made available to the new installation through protected key grants.

## S7 -- Total device loss

Given no trusted installation remains, the user can use independent
high-entropy Recovery Material to restore the Trust Domain and access
the latest recoverable Channel state. Server control alone is
insufficient.

## S8 -- Conveyance unavailable

When Conveyance is unavailable, local Vocation/Illumination operation
remains usable. WGT may continue using the last locally decrypted and
domain-validated snapshot.

## Future scenario -- bidirectional domain synchronization

Vocation or Illumination may later publish domain-owned
change/reconciliation contracts through Conveyance. Conveyance may gain
another generic delivery mode, but it never acquires the foreign
merge/conflict semantics.
