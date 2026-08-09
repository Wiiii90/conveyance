# 09 -- Read Models

Conveyance V1 intentionally has one product-readable delivery projection.

## Current Object

For a route identified by opaque canonical UUID `trust_domain_ref` and
`channel_ref`, an already-authorized caller with `read` permission receives
the current Envelope:

- `envelope_format_version`;
- `channel_epoch`;
- `revision`;
- `envelope_ref`;
- `previous_envelope_ref`;
- base64-encoded opaque `protected_payload`.

Trust Domain and Channel references remain in the route and are not repeated
in the JSON representation. A route with no Current Object returns Not Found;
it does not return an empty synthetic Envelope.

The projection exposes exactly one current Envelope. It has no superseded
Envelope collection, revision list, change feed, or product-visible history.
Replacement may use transactional state internally without expanding this
read model.

The protected payload is returned byte-for-byte after base64 decoding and
encoding boundaries. Conveyance exposes no Vocation, Illumination, WGT, or
cryptographic projection of those bytes.

## Authorization view in v0.1.0

The transport supplies an already-authorized operation context/principal with
generic `read` or `publish` permission. This temporary seam is not a
product-readable installation-access model and defines no authentication
credentials or protocol.

Installation status, grants, public credential metadata, revocation, and Key
Grant projections remain future security-milestone read models.

## Operational state

Server operators may observe request counts, response/error categories,
decoded payload byte sizes, storage consumption, timestamps, and latency.
Server upload timestamps are operational metadata and are not part of the
Current Object JSON contract.

Logs and metrics must not contain protected payload bytes, decrypted foreign
payloads, private keys, Channel Keys, Recovery Material, or raw authentication
secrets.
