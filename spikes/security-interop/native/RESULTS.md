# OpenSSL HPKE bridge candidate results

Date: 2026-08-10  
Repository: `wgt-system/conveyance`  
Worktree: `P:\conveyance\.worktrees\dev`  
Branch: `dev`  
Expected HEAD before this candidate: `48c9fdc3882e518c70c7869c98eb1ff09defcc46`

## Classification

`OPENSSL-HPKE-WINDOWS-PASS`

The previous environment blocker is cleared. NASM 3.02 was available at
`C:\Program Files\NASM\nasm.exe`; x64 MSVC, nmake, and MATLAB Perl were also
available. OpenSSL configured and built with the normal intended path.

## Source and license provenance

- Source: official `openssl-3.5.7.tar.gz` release archive.
- Published SHA-256: `a8c0d28a529ca480f9f36cf5792e2cd21984552a3c8e4aa11a24aa31aeac98e8`.
- Local SHA-256: matched the published value exactly.
- Configure attempted: `VC-WIN64A no-shared no-tests`.
- Exact configure command: `perl Configure VC-WIN64A no-shared no-tests` from
  the extracted OpenSSL 3.5.7 source under the x64 VS developer environment,
  with NASM 3.02 on the run-scoped PATH.
- Build: `nmake` PASS; `libcrypto.lib` produced as a static library.
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

## Native and managed proof

- `OSSL_HPKE_suite_check`: PASS for Base/X25519/HKDF-SHA256/AES-256-GCM.
- Native C shim: PASS; public `<openssl/hpke.h>` and public EVP raw-key APIs
  only. The shim builds warning-free with `/W4 /WX`.
- .NET 10 P/Invoke: PASS; run-scoped resolver loads the shim and no C# crypto
  construction is used.
- Native random round-trip: PASS.
- Selected RFC 9180 vector: PASS, including exact encapsulated key and raw
  32-byte X25519 public-key derivation.
- Failed AAD open: PASS; safe error, zero output length, and cleared plaintext.
- Go -> .NET/OpenSSL grant: PASS.
- .NET/OpenSSL -> Go grant: PASS.
- Grant tamper suite: `6/6 PASS` in both directions.
- Private material: the spike passes raw 32-byte private material through
  managed arrays for this proof; native temporary buffers are cleared on
  failure. Platform-protected/non-exportable HPKE persistence remains
  unresolved and is not production approval.

The following remain `NOT-RUN`:

- real iPhone HPKE interoperability and Keychain-backed storage.

Raw private-key handling and platform-protected persistence remain unresolved
spike questions and are not production approval.

## iOS static audit

`IOS-BRIDGE-PATH-PLAUSIBLE` as a design-only conclusion: the proposed shim
uses a C ABI and public OpenSSL HPKE calls rather than Windows APIs, so the
shape is plausibly portable to an OpenSSL-built Apple target. No shim was
compiled for iOS and no iPhone/device result is claimed. A real Apple build
and interoperability run remains required before any ADR-0007 conclusion.

The prior `BLOCKED-MTLS-WINDOWS-ENVIRONMENT` result is preserved unchanged.
