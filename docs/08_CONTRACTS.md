# 08 -- Contracts

## Current Object HTTP V1

v0.1.0 exposes exactly two Current Object operations:

```text
GET /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current
PUT /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current
```

There is no separate Channel-create endpoint in v0.1.0. The first successful
PUT implicitly establishes the `current_object` Channel and its current
Envelope atomically.

Requests and responses use `application/json`. No history, list, version, or
ordered-delivery endpoint is part of V1.

## Opaque reference encoding

`trust_domain_ref`, `channel_ref`, `envelope_ref`, and a non-null
`previous_envelope_ref` use canonical UUID text:

```text
xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

Canonical output uses lowercase hexadecimal digits and hyphens in those
positions. References are syntactically validated but remain semantically
opaque. Conveyance must not derive a foreign service, capability, object type,
or business meaning from a UUID value.

## Current Object JSON

The PUT request body and successful GET/PUT response body have the same exact
shape:

```json
{
  "envelope_format_version": 1,
  "channel_epoch": 1,
  "revision": 1,
  "envelope_ref": "00000000-0000-0000-0000-000000000001",
  "previous_envelope_ref": null,
  "protected_payload": "AA=="
}
```

The body contains exactly these fields:

- `envelope_format_version`: positive JSON integer; v0.1.0 supports value `1`;
- `channel_epoch`: positive JSON integer;
- `revision`: positive JSON integer;
- `envelope_ref`: canonical UUID text;
- `previous_envelope_ref`: canonical UUID text, nullable only on first
  publish;
- `protected_payload`: standard padded RFC 4648 base64 for opaque bytes.

Missing fields, additional fields, malformed Envelope-reference UUIDs,
malformed base64, wrong JSON types, and non-positive epoch/revision values are
structurally invalid. A malformed route UUID is an invalid reference. The path
supplies `trust_domain_ref` and `channel_ref`; they must not be duplicated in
the JSON body. Server upload time is operational metadata and is not a Current
Object JSON field.

## GET semantics

`GET` requires an already-authorized operation context with `read`
permission.

- `200 OK`: the response body is the exact current Envelope JSON shape above;
- `404 Not Found`: no Current Object exists for the route identity;
- `403 Forbidden`: the supplied operation context lacks `read` permission;
- `503 Service Unavailable`: persistence is unavailable.

GET exposes only the current Envelope. It does not expose superseded state.

## PUT transition semantics

`PUT` requires an already-authorized operation context with `publish`
permission.

First publish is valid only when no Current Object exists and:

- `channel_epoch == 1`;
- `revision == 1`;
- `previous_envelope_ref == null`.

Success returns `201 Created` with the accepted Current Object JSON body and
atomically creates the Channel/current Envelope.

Same-epoch replacement is valid only when:

- `channel_epoch` equals the current epoch;
- `revision` equals the current revision plus one;
- `previous_envelope_ref` equals the current `envelope_ref`.

Epoch advance is valid only when:

- `channel_epoch` equals the current epoch plus one;
- `revision == 1`;
- `previous_envelope_ref` equals the current `envelope_ref`.

Successful same-epoch replacement and epoch advance return `200 OK` with the
accepted Current Object JSON body.

PUT rejects decreasing epochs, skipped epochs, non-consecutive same-epoch
revisions, revision other than one on a new epoch, stale or mismatched
`previous_envelope_ref`, and null `previous_envelope_ref` after first publish.
These order/expected-current failures return `409 Conflict`.

## Atomic compare-and-swap

Validation of the expected current state and installation of its replacement
must be one atomic persistence operation. Competing writers that present the
same previous Envelope reference cannot both succeed. The stale writer
receives `409 Conflict`; silent last-write-wins is forbidden.

An internally transactional persistence design is permitted, but no history
or superseded version is product-readable.

## Payload semantics and limit

Conveyance treats `protected_payload` as opaque bytes. v0.1.0 performs no
Vocation/Illumination/WGT parsing and no cryptographic interpretation.

The maximum payload size is a server technical configuration value with a
v0.1.0 default of 8 MiB (8,388,608 bytes). The limit is applied to decoded
payload bytes, not the base64 text length. A payload over the active limit
returns `413 Content Too Large`.

## Generic error response

Every Current Object error response uses this exact JSON shape:

```json
{
  "error": {
    "code": "conflict"
  }
}
```

The status/code mapping is:

| HTTP status | `error.code` | Meaning |
| --- | --- | --- |
| 400 Bad Request | `invalid_reference` | a route UUID is not canonical UUID text |
| 400 Bad Request | `invalid_envelope` | JSON, Envelope field/reference, base64, or positive-number structure is invalid |
| 403 Forbidden | `forbidden` | already-authorized context lacks the required operation permission |
| 404 Not Found | `current_object_not_found` | GET route has no current object |
| 409 Conflict | `conflict` | first-publish, order, or expected-current/CAS rule failed |
| 413 Content Too Large | `payload_too_large` | decoded payload exceeds the active technical limit |
| 422 Unprocessable Content | `unsupported_envelope_format` | structure is valid but `envelope_format_version` is unsupported |
| 503 Service Unavailable | `unavailable` | persistence cannot serve the operation |

The error shape contains no foreign-domain details. `401 Unauthorized` is
reserved for a future production authentication boundary and is not invented
by the v0.1.0 local/test authorization adapter.

## v0.1.0 authorization seam

Application operations receive an already-authorized operation
context/principal from the transport boundary. The seam conveys only generic
`read` or `publish` permission. v0.1.0 defines no mTLS, bearer token, API key,
custom signature, enrollment, or recovery protocol.

Issue #5 may use an explicitly local/test adapter sufficient to exercise
allowed and forbidden read/publish paths. Production authentication remains
security-milestone work.

## Protection boundary

`envelope_format_version` versions the generic Envelope transport format.
`protected_payload` remains one opaque byte string; any future client-side
protection encoding is contained within those bytes and is not interpreted by
Conveyance.

Security-relevant visible metadata will later be bound by the client-side
protection protocol. Trust Domain and Channel references are taken from the
route rather than duplicated in the JSON body. The cryptographic encoding and
canonical AAD remain gated by ADR-0007 and are not frozen by this contract.

Foreign payload schemas remain owned and versioned by their source bounded
contexts.
