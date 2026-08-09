# ADR-0005 – Separate Installation Authentication, Key Encryption and Hybrid Recovery

Status: Accepted architecture direction; production mechanism subject to ADR-0007 interoperability gate.

## Decision

Installation authentication credentials and key-encryption credentials are separate roles.

V1 prefers TLS 1.3 mutual client authentication over a custom HTTP request-signature protocol, subject to real .NET-for-iOS verification.

Channel Key Grants prefer a standards-based HPKE (RFC 9180) construction. HPKE is not considered implementation-approved until a maintained .NET solution or acceptable native bridge passes Windows/iOS interoperability tests.

Recovery uses independent high-entropy Recovery Material. Server control alone cannot authorize enrollment or decrypt recovery state.

## Consequences

Conveyance stores public credentials, revocation state and opaque grants/packages but no plaintext Channel Keys, client private keys or recovery secret.

Implementation agents may not replace an interoperability failure with an ad-hoc ECDH/KDF/encryption design.
