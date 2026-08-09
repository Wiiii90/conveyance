# ADR-0003 -- Opaque Channel with Current Object as the Only V1 Delivery Mode

Status: Accepted

## Decision

`Channel` is the generic opaque delivery namespace. V1 implements only
`Current Object`: one logical current Envelope per Channel, conditional atomic
replacement, and no product-visible history.

## HTTP V1

v0.1.0 exposes exactly:

```text
GET /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current
PUT /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current
```

There is no separate Channel-create endpoint. The first successful PUT
atomically establishes the `current_object` Channel and current Envelope.

`trust_domain_ref`, `channel_ref`, `envelope_ref`, and non-null
`previous_envelope_ref` use canonical lowercase UUID text on the wire and
remain semantically opaque. Conveyance does not derive foreign meaning from
their values.

## Envelope transport shape

The JSON body contains exactly:

- `envelope_format_version`;
- `channel_epoch`;
- `revision`;
- `envelope_ref`;
- `previous_envelope_ref`, nullable only on first publish;
- `protected_payload`, standard padded base64 for opaque bytes.

Trust Domain and Channel references come from the path and are not duplicated
in the body. v0.1.0 supports Envelope format version 1. A structurally valid
unsupported version returns 422 `unsupported_envelope_format`; structural
Envelope errors return 400 `invalid_envelope`.

## Transition rules

First publish requires no existing Current Object, epoch 1, revision 1, and a
null previous Envelope reference. It returns 201 and creates Channel/current
Envelope atomically.

Same-epoch replacement requires the current epoch, current revision plus one,
and previous Envelope reference equal to the current Envelope reference.

Epoch advance requires current epoch plus one, revision 1, and previous
Envelope reference equal to the current Envelope reference.

Successful replacement returns 200. Decreasing or skipped epochs,
non-consecutive revisions, revision other than one on a new epoch, stale or
mismatched previous Envelope reference, and a null previous Envelope reference
after first publish return 409 Conflict.

Expected-current validation and replacement are one atomic persistence
compare-and-swap operation. Competing stale writers receive Conflict; silent
last-write-wins is forbidden.

GET returns 200 with the exact current Envelope or 404
`current_object_not_found` when none exists for the route.

## Payload and history

Conveyance treats `protected_payload` as opaque decoded bytes. It performs no
foreign-domain parsing and no cryptographic interpretation in v0.1.0.

Maximum payload size is a server technical configuration value. The v0.1.0
default is 8 MiB (8,388,608 decoded bytes), measured after base64 decoding. An
oversized payload returns 413 `payload_too_large`.

Only the current Envelope is product-readable. Persistence may use internal
transactions, but V1 exposes no history, list, version, or ordered-delivery
endpoint.

## Authorization seam

v0.1.0 application operations receive an already-authorized operation
context/principal from the transport boundary with generic `read` or `publish`
permission. Issue #5 may use an explicitly local/test adapter to exercise
permission paths.

This is not an authentication protocol. v0.1.0 defines no mTLS, token, API key,
custom request signature, enrollment, revocation, or recovery behavior.
Production authentication remains security-milestone work.

## Rationale

The contract serves the first PC-off Current Object read flow without defining
Conveyance forever as snapshot-only infrastructure. It gives domain,
persistence, and HTTP implementation tasks one fixed ordering and conflict
model while preserving the opaque bounded-context boundary.

Future ordered/change delivery requires a new justified decision and must not
move Vocation, Illumination, or WGT semantics into Conveyance.

## Consequences

- Issues #3--#5 implement this contract without redesigning it.
- Persistence must provide atomic compare-and-swap and restart durability.
- Duplicate PUT replay after success is stale and returns Conflict; callers
  retrieve current state to reconcile.
- The temporary authorization seam cannot be promoted into production
  authentication.
- Cryptographic format and interoperability remain governed by the security
  ADRs, including ADR-0007.
