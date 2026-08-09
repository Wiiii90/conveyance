# 05 -- Domain Model

## Channel aggregate

`Channel` is the V1 aggregate. Its identity is the pair
`(trust_domain_ref, channel_ref)`. Both references use canonical UUID text at
the transport boundary and remain semantically opaque inside Conveyance.
Conveyance must not infer a foreign service, capability, or business meaning
from either UUID.

V1 has one delivery mode, `current_object`, and stores at most one current
Envelope for a Channel. There is no separate Channel-create operation in
v0.1.0. The first successful publish atomically establishes the Channel and
its current Envelope.

The aggregate's current state is:

- current Channel epoch;
- current revision within that epoch;
- current Envelope reference;
- the immutable current Envelope and its opaque protected payload.

Only the current Envelope is product-readable. An implementation may use
transactional records internally, but no superseded Envelope is part of the
V1 product model or API.

## Envelope

An Envelope is immutable after acceptance. Its transport representation has
exactly these fields:

- `envelope_format_version`;
- `channel_epoch`;
- `revision`;
- `envelope_ref`;
- `previous_envelope_ref`, nullable only for the first publish;
- `protected_payload`, opaque bytes encoded as base64 on the JSON wire.

`trust_domain_ref` and `channel_ref` come from the route and are not duplicated
inside the Envelope JSON. `envelope_ref` and `previous_envelope_ref` use
canonical UUID text and remain semantically opaque.

Conveyance validates generic structure, ordering, version support, and
technical size only. It does not parse or interpret `protected_payload` as a
Vocation, Illumination, WGT, or cryptographic object in v0.1.0.

## Current Object transitions

A first publish is valid only when no Current Object exists and:

- `channel_epoch == 1`;
- `revision == 1`;
- `previous_envelope_ref == null`.

A same-epoch replacement is valid only when:

- `channel_epoch` equals the current epoch;
- `revision` equals the current revision plus one;
- `previous_envelope_ref` equals the current `envelope_ref`.

An epoch advance is valid only when:

- `channel_epoch` equals the current epoch plus one;
- `revision == 1`;
- `previous_envelope_ref` equals the current `envelope_ref`.

The aggregate rejects decreasing or skipped epochs, non-consecutive
same-epoch revisions, a revision other than one on a new epoch, stale or
mismatched `previous_envelope_ref`, and a null `previous_envelope_ref` after
the first publish.

Expected-current validation and replacement form one atomic compare-and-swap
operation. A stale concurrent writer receives Conflict; there is no silent
last-write-wins behavior.

## Authorization seam

v0.1.0 application operations receive an already-authorized operation
context/principal from the transport boundary and enforce generic `read` or
`publish` permission. This seam is not an authentication protocol and does
not define credentials, enrollment, or trust establishment.

## Future security model

Installation Principal, Channel Grant, Key Grant, revocation, and recovery
remain security-milestone concepts. They are distinct from the temporary
v0.1.0 authorization adapter and are not implemented by the Current Object
core milestone.

## Not domain objects

The following are explicitly foreign and must not enter the model:
Opportunity, Company, Posting, LearningItem, Review, Deck, WGT
ServiceIntegration, and the WGT Device aggregate.

## V1 ordering

A Channel Current Object is ordered by `(channel_epoch, revision)`. A new
epoch starts a new revision sequence, and a higher epoch is always newer than
a lower epoch. Client rollback detection additionally uses a local trusted
Checkpoint in the later security design.
