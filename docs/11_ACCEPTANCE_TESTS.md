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

## v0.2.0 security interoperability specification gates

These are specification acceptance cases for the executable spike governed by
ADR-0007 and ADR-0008. They do not authorize production security code.

39. Go uses standard Go 1.26 `crypto/hpke`, the frozen RFC 9180 Base Mode
    suite (`0x0020`, `0x0001`, `0x0002`), and passes the official selected-suite
    vectors.
40. Windows uses a persisted platform installation-authentication credential;
    an actual TLS 1.3 client-authenticated request to Go succeeds, while an
    unregistered or mismatched credential is rejected.
41. A real iPhone .NET runtime uses a Keychain-backed installation identity for
    an actual client-authenticated request to Go; unregistered credentials are
    rejected and simulator-only evidence is not accepted.
42. Windows passes official HPKE vectors, opens Go-produced grants, produces
    grants opened by Go, and rejects all specified tampering.
43. A real iPhone passes official HPKE vectors, opens Go-produced grants,
    produces grants opened by Go, and rejects all specified tampering.
44. Windows produces byte-identical AES-GCM output for the fixed project
    fixture, opens Go output, produces output opened by Go, and rejects all
    specified metadata/nonce/ciphertext/tag tampering.
45. A real iPhone meets the same AES-GCM fixture, round-trip, and tamper
    requirements as Windows.
46. Evidence proves the authentication credential, HPKE credential, and
    Channel Key are separate, and Conveyance never obtains a Channel Key or
    private HPKE key.
47. The spike result is exactly `PASS`, `BLOCKED-HPKE`, `BLOCKED-MTLS-IOS`,
    `BLOCKED-KEY-STORAGE`, or `BLOCKED-OTHER`, with exact evidence. No partial
    result is described as production-ready security.

The required Channel Key Grant and Envelope tamper cases, exact AAD/info
bytes, fixture inputs, and storage evidence are frozen in
`docs/15_SECURITY_INTEROP_PROFILE.md`. Replay remains governed by ADR-0006;
checkpoints are not implemented in this specification issue.

## Deferred security acceptance tests

Production authentication, enrollment, revocation, Key Grant APIs, payload
cryptographic integration, rollback checkpoints, and recovery tests remain
gated security milestone work. They must follow accepted security ADRs and
ADR-0007/ADR-0008 rather than the v0.1.0 local/test authorization adapter.

## Availability tests

37. A Conveyance outage does not make a foreign local application unusable.
38. Previously validated local caches remain usable according to their owning
    bounded-context policy.
