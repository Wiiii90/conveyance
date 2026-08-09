# ADR-0007 – Crypto and mTLS Interoperability Gate

Status: Accepted

## Context

Conveyance server code is planned in Go, while WGT clients are .NET 10/Avalonia on Windows and iPhone.

The security design must therefore be proven across the actual runtime boundary before production code depends on a cryptographic or authentication mechanism.

Current platform facts:

- Go 1.26 provides `crypto/hpke` implementing RFC 9180.
- Go and .NET provide standard AES-GCM primitives.
- .NET 10 provides HTTP/TLS client-certificate APIs.
- .NET for iOS exposes Keychain access, including certificate/key identities.
- Apple CryptoKit provides HPKE, but WGT is C#/.NET, so CryptoKit availability alone does not prove a usable WGT implementation.
- the documented .NET 10 cryptography surface does not provide first-party HPKE.

## Decision

Production security code is blocked until a focused interoperability spike proves the exact client/server combination.

The spike must prove, on real target runtimes where applicable:

1. Windows .NET can generate/store/use an Installation authentication credential.
2. iPhone .NET can generate/store/use an Installation authentication credential from the Keychain and successfully perform the selected client-authenticated TLS request.
3. Go Conveyance can authenticate the enrolled client credential without requiring a server-controlled enrollment authority.
4. Windows .NET and iPhone .NET can create/open the selected Channel Key Grant format.
5. The same RFC/protection test vectors are accepted by both client platforms.
6. AES-256-GCM Envelope vectors with canonical AAD round-trip identically on Windows .NET and iPhone .NET.
7. Tampered AAD/ciphertext/tag fails deterministically.
8. Private key material remains in the platform-appropriate protected store where the selected API permits it.

## HPKE gate

HPKE remains preferred for Channel Key Grants, but it is not implementation-approved merely because Go or Apple exposes HPKE.

Acceptable outcomes are:

- an audited maintained .NET library with RFC 9180 support and successful Windows/iOS tests; or
- a narrowly scoped native-platform bridge that passes the same contract tests and has acceptable packaging/maintenance cost.

If neither is satisfactory, return to the control plane. Do not invent a private hybrid-encryption protocol.

## mTLS gate

mTLS remains the preferred installation-authentication candidate.

The spike must confirm real .NET-for-iOS certificate/key handling and request behavior. If that fails operationally, return to the control plane for another standard authentication mechanism rather than implementing custom request signing ad hoc.

## No implementation shortcut

The server repository may bootstrap and implement non-security Current Object semantics before this gate, but no production claim of secure remote synchronization may be made until the gate passes.
