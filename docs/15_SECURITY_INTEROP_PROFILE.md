# 15 -- v0.2.0 Security Interoperability Profile

Status: Accepted candidate profile for the v0.2.0 interoperability spike.

## Purpose and boundary

v0.2.0 begins with an interoperability proof across Go Conveyance, .NET 10
Windows, and .NET 10 iPhone on a real target runtime. A Windows build of an
iOS target is not sufficient evidence. This profile does not introduce
production authentication endpoints, enrollment, revocation, Recovery
Authority code, Channel Key Grant APIs, production payload-encryption
integration, or WGT/Vocation/Illumination integration.

The Channel Key Grant format and protected-payload framing below are v0.2
interoperability candidates only. They are not published Conveyance HTTP/API
contracts and do not change the v0.1 Current Object contract in
`docs/08_CONTRACTS.md`.

## Credential separation

The following are distinct and must never be reused across roles:

`InstallationRef` != WGT `DeviceIdentity` != Installation Authentication
Credential != Key Encryption Credential != Channel Key != Recovery Material.

The Installation Authentication Credential authenticates an enrolled
installation to Conveyance transport; the candidate is mTLS. The Key
Encryption Credential receives Channel Key Grants; the candidate is an HPKE
recipient key. The mTLS authentication key must never be used as the HPKE
recipient key. A Channel Key is exactly 32 random bytes, scoped to one Channel
epoch, and encrypts opaque protected Envelope payloads. Recovery Material is
separate and is not implemented in this spike.

## mTLS candidate

The transport candidate is TLS 1.3 with mutual TLS. The spike certificate is
X.509 v3 with an ECDSA P-256 authentication key, SHA-256 certificate
signature, Key Usage including Digital Signature, and Extended Key Usage
`1.3.6.1.5.5.7.3.2` (TLS Web Client Authentication). Each Installation has a
distinct certificate/key.

InstallationRef is never inferred from certificate CN, SAN, hostname,
WGT DeviceIdentity, or IP address. The test server uses an explicitly
pre-enrolled mapping:

```text
(trust_domain_ref, installation_ref)
    -> SHA-256 fingerprint of the installation authentication SPKI
```

This mapping represents completed explicit enrollment for the fixture. It is
not the production enrollment protocol, and changing server-owned data alone
must not be treated as sufficient future trust establishment. Test-only
CA/server certificates and localhost topology are sufficient; certificate
lifetime and renewal policy remain undecided.

The Windows gate requires a persisted installation authentication key,
platform certificate/private-key storage where supported, and an actual
client-authenticated TLS request to Go without plaintext private-key
persistence. The iPhone gate requires a real .NET/iOS target, an
iOS-Keychain-backed identity, and an actual client-authenticated request to
Go. Simulator or shared-project compilation alone is not evidence. If the
Keychain-backed identity cannot be used without unacceptable private-key
export, the mTLS gate fails and returns to the Control Plane.

## HPKE candidate

Use RFC 9180 HPKE Base Mode only. The frozen suite is:

| Component | Value | ID |
| --- | --- | --- |
| KEM | DHKEM(X25519, HKDF-SHA256) | `0x0020` / 32 |
| KDF | HKDF-SHA256 | `0x0001` / 1 |
| AEAD | AES-256-GCM | `0x0002` / 2 |

The Go reference uses standard Go 1.26 `crypto/hpke` APIs. No custom ECDH,
HKDF composition, HPKE, or private hybrid construction is permitted. HPKE
Auth mode is not used because installation transport authentication belongs
to mTLS.

The .NET candidate must be maintained, RFC 9180-conformant, support the exact
suite on Windows .NET 10 and real `net10.0-ios`, have an acceptable license,
and pass RFC test-vector compatibility. Compilation is not approval. If no
acceptable implementation or narrowly scoped native bridge passes both
platform gates, the result is `BLOCKED-HPKE`; homemade ECDH+HKDF+AES is not a
fallback.

Each Installation has a separate HPKE recipient credential. Its public key
uses RFC 9180 `SerializePublicKey`: exactly 32 X25519 bytes, represented in
JSON/test fixtures as standard padded RFC 4648 Base64. Private keys never
appear in production wire data or belong to Conveyance merely because it
stores grants. Deterministic private material is allowed only in clearly
marked testdata. The spike must report private-key persistence, managed-memory
materialization, and whether platform-protected/non-exportable storage is
possible for the selected implementation.

## Channel Key Grant 1.0 candidate

This exact JSON shape has no extra fields and is not yet a public HTTP
contract:

```json
{
  "grant_format_version": 1,
  "kem_id": 32,
  "kdf_id": 1,
  "aead_id": 2,
  "trust_domain_ref": "<canonical-lowercase-uuid>",
  "channel_ref": "<canonical-lowercase-uuid>",
  "channel_epoch": 7,
  "recipient_installation_ref": "<canonical-lowercase-uuid>",
  "enc": "<standard-padded-base64>",
  "ciphertext": "<standard-padded-base64>"
}
```

`enc` is raw RFC 9180 encapsulated-key bytes, Base64 only for JSON.
`ciphertext` is the HPKE ciphertext including its AEAD tag, Base64 only for
JSON. The plaintext is exactly the 32 raw Channel Key bytes, with no JSON
wrapper. HPKE info is the exact UTF-8 bytes:

```text
conveyance/channel-key-grant/1.0
```

There is no NUL or trailing newline. HPKE AAD is the UTF-8 encoding of these
LF-separated lines, with no trailing LF:

```text
conveyance-channel-key-grant-v1
trust_domain_ref=<trust_domain_ref>
channel_ref=<channel_ref>
channel_epoch=<channel_epoch>
recipient_installation_ref=<recipient_installation_ref>
```

UUIDs are canonical lowercase textual UUIDs. `channel_epoch` is unsigned
base-10 ASCII, positive, with no leading plus or leading zero. Go must produce
grants opened by Windows and iPhone .NET, and each .NET target must produce a
grant opened by Go. Each implementation also runs the relevant official RFC
9180 vectors.

## Envelope AES-256-GCM candidate

This candidate does not change the v0.1 opaque `protected_payload` contract.
Using the 32-byte Channel Key, AES-256-GCM uses a 12-byte nonce and 16-byte
authentication tag. The binary payload is exactly:

```text
nonce || ciphertext || tag
```

The complete bytes remain Base64-encoded once by the existing Current Object
JSON. There is no nested Base64 or embedded JSON crypto wrapper. Production
nonce generation and the per-epoch uniqueness decision are deferred; the
spike uses fixed deterministic nonces only as test vectors.

Envelope AAD is UTF-8, LF-separated, and has no trailing LF:

```text
conveyance-current-object-envelope-v1
trust_domain_ref=<trust_domain_ref>
channel_ref=<channel_ref>
envelope_format_version=<envelope_format_version>
channel_epoch=<channel_epoch>
revision=<revision>
envelope_ref=<envelope_ref>
previous_envelope_ref=<previous_envelope_ref-or-null>
```

UUIDs are canonical lowercase. Integer fields are unsigned base-10 ASCII
without a plus sign or leading zeroes for positive values. Absent
`previous_envelope_ref` is literal ASCII `null`. These fields bind route,
format, epoch, revision, and both Envelope identities. Any alteration must
fail authentication.

## Deterministic project fixture

```text
trust_domain_ref:              00000000-0000-0000-0000-000000000101
channel_ref:                   00000000-0000-0000-0000-000000000102
recipient_installation_ref:    00000000-0000-0000-0000-000000000103
previous_envelope_ref:         00000000-0000-0000-0000-000000000104
envelope_ref:                  00000000-0000-0000-0000-000000000105
envelope_format_version:       1
channel_epoch:                 7
revision:                      11
```

Channel Key bytes, hex:

```text
000102030405060708090a0b0c0d0e0f
101112131415161718191a1b1c1d1e1f
```

AES-GCM nonce bytes, hex: `a0a1a2a3a4a5a6a7a8a9aaab`

Plaintext is the exact UTF-8 bytes of
`{"fixture":"conveyance-security-interop-v1"}`, with no trailing newline.
This specification deliberately does not invent expected ciphertext/tag
bytes. Independent Go and .NET implementations must produce identical output
before a canonical expected fixture is committed by the implementation issue.

## Tamper, replay, and compromise boundaries

The executable spike must show deterministic failure, with no accepted or
partially surfaced plaintext, for altered grant route/epoch/recipient values,
`enc`, or ciphertext/tag; and for altered Envelope route/format/epoch/revision/
identity values, nonce, ciphertext, or tag.

AEAD does not prevent replay. ADR-0006 remains authoritative: trusted client
checkpoints contain `(channel_epoch, revision, envelope_ref)`, and rollback
rejection is a separate client rule. Checkpoints are not implemented here.

Conveyance may know generic references, epochs, revisions, grant/routing
metadata, and encrypted bytes. It must not require foreign plaintext, Channel
Keys, HPKE recipient private keys, or Recovery private material. The
pre-enrolled fingerprint map replaces completed enrollment in the fixture; it
is not a server-controlled trust-establishment mechanism.

## Result classification

The spike result is exactly one of:

- `PASS`: every required Go, Windows, and real-iPhone gate is proven.
- `BLOCKED-HPKE`: no acceptable maintained .NET/iOS HPKE implementation or
  bridge passes.
- `BLOCKED-MTLS-IOS`: a Keychain-backed .NET/iOS identity cannot perform the
  accepted mTLS request.
- `BLOCKED-KEY-STORAGE`: the candidate requires unacceptable persistent
  plaintext/private-key handling.
- `BLOCKED-OTHER`: exact evidence is documented and returned to the Control
  Plane.

No partial result is production-ready security. Enrollment, recovery,
revocation, production nonce policy, and production security implementation
scope remain deferred until Control-Plane review after the result.
