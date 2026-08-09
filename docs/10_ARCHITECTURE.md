# 10 – Architecture

## Status

Accepted architecture baseline for specification-first Conveyance V1.

Production security implementation remains gated by the interoperability proof in ADR-0007.

## Runtime

Preferred server stack:

- Go;
- HTTP service;
- durable server-side persistence;
- container-friendly single deployable;
- no internal microservice split.

Go 1.26 includes the RFC 9180 `crypto/hpke` standard-library package. This makes HPKE straightforward on the Conveyance/Go side, but the server normally stores opaque Key Grants rather than decrypting foreign business payloads.

Exact Go package layout remains an implementation decision. Issue #4 selects and documents the minimal durable persistence technology consistent with the frozen Current Object contract; it must stop if that selection requires a new architecture decision.

## v0.1.0 Current Object boundary

v0.1.0 exposes only:

- `GET /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current`;
- `PUT /v1/trust-domains/{trust_domain_ref}/channels/{channel_ref}/current`.

There is no separate Channel-create endpoint. The first valid PUT atomically
establishes the `current_object` Channel and its current Envelope. Trust
Domain, Channel, and Envelope references use canonical UUID text and remain
semantically opaque.

The Current Object JSON body carries Envelope format version, epoch, revision,
Envelope reference, nullable-on-first-publish previous Envelope reference, and
base64-encoded opaque protected payload. Trust Domain and Channel references
come from the route and are not duplicated in the body.

Only the current Envelope is product-readable. No history, list, version, or
ordered-delivery endpoint exists in V1.

## Current Object consistency boundary

First publish, same-epoch replacement, and epoch advance follow the exact
transition rules frozen in ADR-0003 and `docs/08_CONTRACTS.md`. Expected-current
validation and replacement must be one atomic persistence compare-and-swap.
A stale concurrent writer receives Conflict; last-write-wins is not permitted.

The protected payload remains opaque. Its configured technical size limit is
checked on decoded bytes and defaults to 8 MiB in v0.1.0.

## v0.1.0 authorization seam

The transport boundary supplies application operations with an
already-authorized operation context/principal carrying generic `read` or
`publish` permission. An explicitly local/test adapter may exercise these
permission paths in Issue #5.

This seam is not an authentication protocol. v0.1.0 defines no mTLS, token,
API-key, custom-signature, enrollment, or recovery mechanism. Production
authentication remains gated security-milestone work.

## Security layers

Five responsibilities remain distinct:

1. server identity;
2. installation authentication;
3. Channel authorization;
4. Channel-key possession/key-encryption capability;
5. foreign business-payload decryption capability.

Server administration must not imply enrollment authority, Channel-key possession, or business-payload decryption capability.

## Trust ownership boundary

Conveyance stores the server-side access-control projection required for delivery:

- opaque Trust Domain reference;
- Installation principal reference;
- public authentication credential material;
- active/revoked state;
- generic Channel permissions;
- opaque Key Grants.

Conveyance is **not** the authority that decides personal trust.

Enrollment approval and emergency recovery authority remain client controlled under the WGT personal-trust architecture. A compromised server can deny service or falsify its own access-control database, but it cannot mint the client-held secrets required to decrypt new Channel epochs.

## Transport authentication

V1 candidate: TLS 1.3 with installation-specific mutual client authentication rather than a custom signed-request protocol.

.NET 10 exposes client-certificate authentication through `HttpClientHandler` / `SslClientAuthenticationOptions`, and .NET for iOS exposes access to Keychain identities. This makes mTLS technically plausible for WGT clients, but real iPhone certificate creation/storage/use must be proven by the ADR-0007 smoke test before this becomes an implementation commitment.

Installation authentication credentials are separate from key-encryption credentials.

## Channel Key hierarchy

Each Channel Epoch has a fresh 256-bit Channel Key.

Channel Keys are available only to authorized trusted clients through protected Key Grants.

HPKE (RFC 9180) remains the preferred standards-based abstraction for Key Grants because:

- Go 1.26 has a standard `crypto/hpke` implementation;
- Apple CryptoKit exposes HPKE;
- it avoids a custom ECDH/KDF/wrap protocol.

However, WGT is .NET/C#, not Swift. .NET 10 does not expose a first-party HPKE API in its documented cryptography surface. Therefore an audited .NET implementation or a narrowly justified native bridge must be proven before HPKE is frozen as the production grant mechanism.

No custom “ECDH + hash + AES” construction may be substituted merely to avoid this gate.

## Envelope payload protection

The previous draft's proposed custom per-envelope KDF construction is **not accepted**.

V1 payload protection should use a directly interoperable, standard AEAD construction. The preferred candidate for the interoperability spike is:

- AES-256-GCM;
- 256-bit Channel Key;
- fresh random 96-bit nonce per Envelope;
- 128-bit authentication tag;
- canonical Conveyance security metadata as AAD.

AES-GCM is available in .NET 10, Go, and Apple CryptoKit.

The protocol must cap the number of encryptions under one Channel Key/epoch to a conservative value below the standard random-nonce safety bound and must rotate the epoch before that bound can be approached. The personal V1 workload is orders of magnitude below that limit.

The concrete wire encoding of nonce, tag, ciphertext, and AAD is frozen only after cross-platform test vectors pass.

## Authenticated metadata

At minimum the following are cryptographically bound as AAD:

- envelope format version;
- Trust Domain reference;
- Channel reference;
- Channel epoch;
- revision;
- Envelope reference;
- previous/expected Envelope reference where present.

Conveyance may read these generic values but cannot alter them without client verification failure.

## Revocation

Revocation has two separate effects:

1. **transport access:** revoked installation credentials are rejected;
2. **payload confidentiality:** active synchronized Channels advance to a new epoch with a fresh Channel Key before confidential future content is published.

Old plaintext/key knowledge cannot be revoked retroactively.

Conveyance never performs server-side business-payload re-encryption. A trusted/recovered client decrypts the old current state if necessary and republishes it under the new epoch.

## Recovery

Recovery uses independent high-entropy Recovery Material, not server control and not a low-entropy memorized password as the sole root.

An opaque encrypted Recovery Package may be stored by Conveyance and contains sufficient protected state to recover:

- Trust Domain recovery state;
- current Channel-key state required for existing synchronized data;
- trusted rollback/checkpoint information to the extent preserved by the recovery design.

Recovery freshness after total loss cannot be guaranteed against a fully malicious rollback server without an independent trusted checkpoint/transparency mechanism. V1 explicitly accepts this limitation.

## Availability

Conveyance is optional infrastructure.

Local WGT, Vocation, and Illumination operation remains possible when Conveyance is absent or unavailable. A Conveyance outage is not equivalent to foreign-domain unavailability.

## Future delivery modes

The Channel abstraction does not equate Conveyance with Current Object forever.

Only `current_object` is implemented in V1. Future ordered/change delivery requires a separate justified contract and must not move Vocation/Illumination reconciliation semantics into Conveyance.
