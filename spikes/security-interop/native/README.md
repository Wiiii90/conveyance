# OpenSSL HPKE native bridge candidate

This directory contains the Windows build gate for the minimal native HPKE
bridge candidate required by ADR-0007. The candidate is deliberately kept
inside the security interoperability spike; it is not a production runtime,
contract, or dependency.

The intended bridge is a tiny C ABI over the public OpenSSL HPKE API in
`<openssl/hpke.h>`, consumed by a run-scoped .NET 10 P/Invoke harness. It will
use RFC 9180 Base mode with X25519 (`0x0020`), HKDF-SHA256 (`0x0001`), and
AES-256-GCM (`0x0002`). No cryptographic construction is implemented here.

Run `run-windows.ps1` from this directory. The script downloads only the
official OpenSSL 3.5.7 source archive when needed, verifies its published
SHA-256, discovers the x64 build tools, and stops at the first missing
prerequisite. Archives, source trees, build products, and native binaries are
run-scoped and must not be committed.

The current result is recorded in `RESULTS.md`.
