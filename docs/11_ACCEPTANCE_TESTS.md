# 11 -- Acceptance Tests

## Boundary tests

1. Server persistence contains opaque protected payload bytes and no
   Conveyance-decoded plaintext foreign business payload.
2. Conveyance code has no dependency on Vocation, Illumination, or WGT domain
   models.
3. Unknown protected payload bytes can be stored and retrieved without
   interpretation.
4. Trust Domain, Channel, and Envelope UUID references remain semantically
   opaque.

## Transport contract tests

5. GET and PUT are available only at
   `/v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current`.
6. No separate Channel-create, history, list, version, or ordered-delivery
   endpoint exists.
7. Route and Envelope references require canonical UUID text.
8. Current Object JSON contains exactly the frozen six fields and does not
   duplicate Trust Domain or Channel references.
9. A malformed route UUID returns 400 `invalid_reference`.
10. Malformed JSON, Envelope-reference UUID, base64, field types,
    missing/extra fields, or non-positive epoch/revision returns 400
    `invalid_envelope`.
11. An otherwise valid unsupported `envelope_format_version` returns 422
    `unsupported_envelope_format`.
12. Every error uses the frozen generic error JSON shape and contains no
    foreign-domain details.

## Current Object transition tests

13. First publish at epoch 1/revision 1 with null previous Envelope reference
    atomically creates the Channel/current Envelope and returns 201.
14. First publish rejects any other epoch, revision, or non-null previous
    Envelope reference with 409 Conflict.
15. Same-epoch replacement accepts only the next consecutive revision and the
    exact current Envelope reference, then returns 200.
16. Epoch advance accepts only the next consecutive epoch at revision 1 with
    the exact current Envelope reference, then returns 200.
17. Decreasing or skipped epochs are rejected with 409 Conflict.
18. Non-consecutive same-epoch revision and revision other than one on a new
    epoch are rejected with 409 Conflict.
19. Null, stale, or mismatched previous Envelope reference after first publish
    is rejected with 409 Conflict.
20. Successful GET returns 200 and the exact current Envelope.
21. GET for a route with no Current Object returns 404
    `current_object_not_found`.
22. A replay after a successful PUT is stale and returns 409 rather than
    creating another revision.

## Persistence tests

23. Expected-current validation and replacement occur in one atomic
    persistence operation.
24. Two competing replacements from the same current Envelope cannot both
    succeed; the stale writer receives 409 Conflict.
25. Successful state survives repository/process restart.
26. No product API exposes superseded Envelope history.
27. Persistence unavailability maps to 503 `unavailable` without exposing
    storage details.

## Payload tests

28. Payload size is measured after base64 decoding.
29. The default v0.1.0 limit accepts payloads up to 8,388,608 decoded bytes.
30. A payload above the active technical limit returns 413
    `payload_too_large`.
31. Accepted opaque bytes round-trip exactly.

## v0.1.0 authorization-seam tests

32. An operation context with `read` permission can retrieve Current Object.
33. A context without `read` receives 403 `forbidden`.
34. An operation context with `publish` permission can attempt first publish
    or replacement.
35. A context without `publish` receives 403 `forbidden`.
36. The local/test adapter introduces no mTLS, token, API-key, custom-signature,
    enrollment, or trust-establishment protocol.

## Deferred security acceptance tests

Production authentication, revocation, Key Grants, payload cryptographic
verification, rollback checkpoints, and recovery tests remain gated security
milestone work. They must follow accepted security ADRs and ADR-0007 rather
than the v0.1.0 local/test authorization adapter.

## Availability tests

37. A Conveyance outage does not make a foreign local application unusable.
38. Previously validated local caches remain usable according to their owning
    bounded-context policy.
