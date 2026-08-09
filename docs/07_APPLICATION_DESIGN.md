# 07 -- Application Design

## V1 use cases

### Register/maintain authenticated installation state

Maintain server-side public credential/status records required to
authenticate already-authorized installations.

This is not permission for the server to create trust unilaterally.

### Create/configure Channel

Create an opaque Channel in a Trust Domain with V1 delivery mode
`current_object`.

### Grant Channel access

Associate an active installation with generic `read` and/or `publish`
permission.

### Store Key Grant

Persist opaque protected Channel-key material for an installation.

### Publish Current Object

Input includes: - Channel identity; - expected current state; -
epoch/revision; - immutable Envelope; - protected payload.

Behavior: - authenticate installation; - authorize `publish`; - validate
Conveyance envelope structure/version/limits; - enforce epoch/revision
and compare-and-swap invariants; - atomically replace current Envelope.

### Retrieve Current Object

Behavior: - authenticate installation; - authorize `read`; - return
current opaque Envelope or explicit absence.

### Revoke Installation

Reject future authenticated operations for the revoked installation.

Client-controlled rekey creates fresh Channel epochs/keys. Conveyance
never derives or re-encrypts business payloads.

### Recovery-package storage

Store/retrieve one opaque current recovery package for the Trust Domain.
Its plaintext and recovery secret are never available to Conveyance.

## Transaction boundary

Current Object replacement must be atomic with its expected-current
check.

## Idempotency

Exact API idempotency semantics are deferred to contract design, but
duplicate retries must not silently create divergent revisions.
