# 08 -- Contracts

This document freezes semantics, not final HTTP paths or JSON field
encoding.

## Conveyance API contract families

V1 requires versioned contracts for:

1.  Installation authentication/registration state required after
    client-authorized enrollment.
2.  Channel creation/configuration.
3.  Channel authorization grants.
4.  Channel Key Grants.
5.  Current Object publish.
6.  Current Object retrieval.
7.  Installation revocation.
8.  Current Recovery Package storage/retrieval.

## Minimal visible Envelope metadata

-   envelope format version;
-   trust domain reference;
-   channel reference;
-   channel epoch;
-   revision;
-   envelope reference;
-   previous/expected envelope reference where required;
-   ciphertext/protected-payload length;
-   server upload timestamp as operational metadata.

Foreign service/capability names are not required server-visible
metadata.

Specifically not visible/required: - `vocation.opportunity_overview`; -
Vocation `publication_ref`; - Vocation `generated_at`; -
Opportunity/Company references; - Illumination learning semantics.

## Authenticated associated data

Security-relevant visible Envelope metadata must be cryptographically
bound to the protected payload as AEAD associated authenticated data
using one canonical encoding defined by the protocol contract.

At minimum: - envelope format version; - trust domain reference; -
channel reference; - channel epoch; - revision; - envelope reference; -
previous envelope reference when present.

## Generic outcomes

-   Unauthorized
-   Forbidden
-   Route/Channel Not Found
-   Current Object Not Found
-   Conflict
-   Payload Too Large
-   Invalid Envelope
-   Unsupported Envelope Version
-   Unavailable

No foreign-domain error semantics are permitted.

## Contract-version rule

HTTP/OpenAPI may describe the Conveyance transport API, but foreign
payload schemas remain owned and versioned by their source bounded
contexts.


## Protection profile

The Envelope contract carries versioned protection metadata sufficient for the client to select the accepted protection profile.

The initial interoperability candidate is AES-256-GCM with a fresh random 96-bit nonce and 128-bit tag. Exact field encoding is not frozen until ADR-0007 test vectors pass.

Algorithm negotiation is not dynamic in V1. A supported Envelope/protection version selects a fixed accepted profile; this prevents downgrade-by-negotiation complexity.
