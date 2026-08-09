# 07 -- Application Design

## v0.1.0 operation boundary

The Current Object core exposes application operations independent of HTTP
and persistence implementation details. Each operation receives:

- opaque `trust_domain_ref` and `channel_ref` values;
- an already-authorized operation context/principal supplied by the transport
  boundary;
- the operation-specific input.

The context carries only the generic permission needed by the operation:
`read` or `publish`. It does not carry or validate mTLS credentials, tokens,
API keys, enrollment proofs, Channel Keys, or Recovery Material.

Issue #5 may provide an explicitly local/test adapter that supplies this
context so permission paths can be exercised. Production authentication is
security-milestone work.

## Retrieve Current Object

Input:

- route identity `(trust_domain_ref, channel_ref)`;
- an operation context with `read` permission.

Behavior:

1. enforce the generic `read` permission;
2. retrieve the current Envelope for the route identity;
3. return the exact opaque Envelope, or Current Object Not Found when none
   exists.

No history or foreign payload projection is available.

## Publish Current Object

Input:

- route identity `(trust_domain_ref, channel_ref)`;
- an operation context with `publish` permission;
- an immutable Envelope containing format version, epoch, revision, Envelope
  references, and opaque protected payload.

Behavior:

1. enforce the generic `publish` permission;
2. validate reference and Envelope structure;
3. reject unsupported Envelope format versions explicitly;
4. enforce the configured decoded-payload limit, whose v0.1.0 default is
   8 MiB;
5. atomically validate first-publish or replacement ordering against the
   current state and install the replacement;
6. report whether the result was a first publish or replacement.

There is no separate Channel-create operation. A valid first publish creates
the `current_object` Channel and current Envelope in the same atomic
persistence operation.

## Transition rules

- First publish: epoch 1, revision 1, null previous Envelope reference.
- Same-epoch replacement: unchanged epoch, next consecutive revision, and
  previous Envelope reference equal to the current Envelope reference.
- Epoch advance: next consecutive epoch, revision 1, and previous Envelope
  reference equal to the current Envelope reference.

All decreasing/skipped epochs, invalid revision transitions, missing or stale
expected-current references, and null previous references after first publish
produce Conflict. Structurally malformed inputs produce Invalid Envelope.

## Persistence port and transaction boundary

The application persistence port must make expected-current validation and
replacement one atomic compare-and-swap operation. It must distinguish:

- first publish accepted;
- replacement accepted;
- conflict because current state did not match;
- current object absent on read;
- storage unavailable.

The persistence technology is selected in Issue #4. Selection must not alter
these semantics or expose product-visible history.

## Retry behavior

Every PUT is a compare-and-swap attempt. Replaying a request after its first
successful application finds a different current state and returns Conflict;
the caller retrieves current state to reconcile. A retry never silently
creates a divergent revision and no last-write-wins path exists.

## Deferred operations

Production installation authentication, Channel grants, Key Grants,
revocation, recovery-package storage, and cryptographic processing remain
outside v0.1.0. Ordered delivery is also deferred.
