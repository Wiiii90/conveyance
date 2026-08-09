# 05 -- Domain Model

## Aggregate direction

### Channel

Primary V1 aggregate candidate.

Identity: - `trust_domain_ref` - `channel_ref`

State: - delivery mode (`current_object` in V1); - current epoch; -
current revision; - current envelope reference; - authorization/grant
associations as application/security policy permits.

Invariants: - Channel references are opaque. - Epoch never decreases. -
Revision never decreases within an epoch. - Current Object replacement
is conditional on the caller's expected current Envelope/state. - V1
exposes no product-visible Envelope history.

### Envelope

Immutable value/entity associated with a Channel revision.

Contains Conveyance metadata and opaque protected payload. It contains
no foreign domain object model.

### Installation Principal

Security-facing identity known to Conveyance: - `installation_ref`; -
`trust_domain_ref`; - authentication public credential/state; - status
(`active`, `revoked`).

It is not WGT's Device aggregate.

### Channel Grant

Generic authorization relation: - installation; - channel; - `read`
and/or `publish`.

### Key Grant

Opaque protected Channel-key material targeted to an installation.
Conveyance stores it but cannot recover the plaintext Channel Key.

## Not domain objects

The following are explicitly foreign and must not enter the model:
Opportunity, Company, Posting, LearningItem, Review, Deck, WGT
ServiceIntegration, WGT Device aggregate.

## V1 ordering

A Channel Current Object is ordered by `(channel_epoch, revision)`.

A new epoch starts a new revision sequence. A higher epoch is always
newer than a lower epoch.

Client rollback detection additionally uses a local trusted Checkpoint.
