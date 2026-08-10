# OpenSSL HPKE bridge candidate results

Date: 2026-08-10  
Repository: `wgt-system/conveyance`  
Worktree: `P:\conveyance\.worktrees\dev`  
Branch: `dev`  
Expected HEAD before this candidate: `48c9fdc3882e518c70c7869c98eb1ff09defcc46`

## Classification

`BLOCKED-OPENSSL-BUILD-ENVIRONMENT`

The official OpenSSL 3.5.7 source was verified, and its public HPKE header
contains the required API and suite identifiers. The x64 MSVC, nmake, and a
Perl installation with the required OpenSSL module were available. OpenSSL
configuration could not produce a build because NASM was not installed or
discoverable:

`NASM not found - make sure it's installed and available on %PATH%`

No NASM installation or substitute assembler was installed. This is an
environment gate, not an OpenSSL HPKE API rejection.

## Source and license provenance

- Source: official `openssl-3.5.7.tar.gz` release archive.
- Published SHA-256: `a8c0d28a529ca480f9f36cf5792e2cd21984552a3c8e4aa11a24aa31aeac98e8`.
- Local SHA-256: matched the published value exactly.
- Configure attempted: `VC-WIN64A no-shared no-tests`.
- OpenSSL 3.5 is an LTS line; the official source page lists 3.5.7 and its
  LTS/EOL information. The project is Apache License 2.0 for this release.
- No archive, source tree, build output, library, or executable was added to
  the repository.

## API and suite audit

The official 3.5.7 `include/openssl/hpke.h` declares the required public
operations, including context creation/free, suite checking, key generation,
encapsulation/decapsulation, seal/open, and deterministic input-key material
where supported. The header defines the frozen suite IDs: X25519 `0x0020`,
HKDF-SHA256 `0x0001`, and AES-256-GCM `0x0002`.

The following remain `NOT-RUN` because no native library could be built:

- native suite check, random round-trip, and AAD tamper proof;
- selected RFC 9180 vector open verification;
- .NET 10 run-scoped native loading and P/Invoke proof;
- Go-to-OpenSSL and OpenSSL-to-Go grant interop and six-case tamper suite;
- private-key input, error-safety, and failed-open plaintext checks.

Raw private-key handling and platform-protected persistence remain unresolved
spike questions and are not production approval.

## iOS static audit

`IOS-BRIDGE-PATH-PLAUSIBLE` as a design-only conclusion: the proposed shim
uses a C ABI and public OpenSSL HPKE calls rather than Windows APIs, so the
shape is plausibly portable to an OpenSSL-built Apple target. No shim was
compiled for iOS and no iPhone/device result is claimed. A real Apple build
and interoperability run remains required before any ADR-0007 conclusion.

The prior `BLOCKED-MTLS-WINDOWS-ENVIRONMENT` result is preserved unchanged.
