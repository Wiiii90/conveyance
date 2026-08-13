# v0.2.0 Go + Windows spike results

Run date: 2026-08-10

Canonical repository: `wgt-system/conveyance`.
Canonical spike Go module: `github.com/wgt-system/conveyance/spikes/security-interop/go`.

This is empirical spike evidence only. It is not ADR-0007 `PASS`, not
production-security approval, and does not prove the real-iPhone gate.

## Stage conclusion

`BLOCKED-HPKE`

No acceptable maintained managed C# implementation of RFC 9180 HPKE with the
frozen Base-mode X25519/HKDF-SHA256/AES-256-GCM suite was identified. No native
bridge was created. The next permitted step is Control-Plane review and a
later narrowly scoped candidate/bridge investigation; homemade ECDH+HKDF+AES
is explicitly rejected.

## Toolchains

- Go: `go1.26.5 windows/amd64` (`C:\Program Files\Go\bin\go.exe`).
- .NET SDK: `10.0.302`; runtime `10.0.10`.
- OS: Windows 10.0.26200, win-x64.
- .NET iOS workload is installed, but no physical iPhone was used in this
  Windows half.

## Go reference

- Uses only standard-library `crypto/hpke`, `crypto/ecdh`, `crypto/aes`,
  `crypto/cipher`, `crypto/tls`, `crypto/x509`, and `crypto/sha256`.
- Frozen suite: DHKEM(X25519, HKDF-SHA256), HKDF-SHA256, AES-256-GCM.
- Normal HPKE Channel Key Grant round trip: `PASS`.
- Grant AAD tamper rejection: `PASS`.
- Go 1.26.5 `crypto/hpke/testdata/rfc9180.json` entry for mode 0/KEM 0x0020/
  KDF 0x0001/AEAD 0x0002 is the exact provenance of the deterministic
  material. It is RFC 9180-derived corpus material, not an RFC Appendix A
  vector: Appendix A.1 uses AES-128-GCM and does not define this frozen
  AES-256-GCM combination.
- RFC 9180 algorithm/suite conformance and the frozen-suite deterministic
  cross-runtime proof are reported separately in the native results.
- Same-suite caller-supplied info/AAD round trip: `PASS`.
- The installed standard-library testdata is compact and stores accumulated
  values. Its deterministic sender test hook is internal; the public API does
  not permit reproducing the fixed ephemeral sender. No internal/unsafe API was
  used. This limitation is recorded as `UNAVAILABLE_PUBLIC_API`; no false
  full-vector-open claim is made.

## Windows mTLS credential

- Implementation uses .NET 10 `CngKey.Create` with ECDSA P-256, the Microsoft
  Software Key Storage Provider, a unique persisted key name, signing usage,
  and `CngExportPolicies.None`.
- Certificate profile: self-signed test X.509 v3, SHA-256, Digital Signature,
  TLS Web Client Authentication EKU `1.3.6.1.5.5.7.3.2`; public certificate
  only is written to the run directory.
- Diagnostic result: ephemeral ECDSA P-256 CNG creation and signing `PASS`.
- Microsoft Software Key Storage Provider is the selected provider, but the
  minimal uniquely named persisted CurrentUser key creation fails:
  `System.Security.Cryptography.CryptographicException`, HRESULT
  `0x80070002`, message `The system cannot find the file specified.`
- The first failure is persisted minimal key creation. Persisted signing,
  persisted `CngExportPolicies.None`, and separate-process reopen are
  `NOT-RUN` because no persisted key exists. The diagnostic cleanup is
  run-scoped and idempotent.
- mTLS sub-gate: `BLOCKED-MTLS-WINDOWS-ENVIRONMENT`. This is host/store access
  evidence, not an architectural mTLS failure. The client code does not fall
  back to an exportable or in-memory credential.
- Because persisted credential creation did not become executable, no-client,
  registered, unknown-certificate, mismatched-InstallationRef, and negotiated
  TLS 1.3 request cases are `NOT-RUN` in this environment. The harness now
  contains all four request paths and will execute them if the diagnostic
  prerequisite passes.
- Server implementation is loopback-only, TLS 1.3-only, requires a client
  certificate, pins the test server SPKI in the client, and checks the explicit
  InstallationRef mapping. No trust-all callback or production CA is used.

## .NET HPKE candidate audit

### OpenSSL 3.5.7 native bridge candidate

The official OpenSSL 3.5.7 source archive and published SHA-256 were
verified. The public `openssl/hpke.h` header contains the frozen RFC 9180
suite identifiers and the required context, key, encapsulation, seal/open,
and suite-check operations. The normal x64 build produced static `libcrypto`;
the candidate result is `OPENSSL-HPKE-WINDOWS-PASS`.

The public-API C shim and .NET 10 P/Invoke proof passed RFC 9180
algorithm/suite conformance plus the separate Go-derived frozen-suite
deterministic proof, native random/tamper proof, both Go↔OpenSSL grant
directions, and exactly these six tamper cases in both directions:
`trust_domain_ref`, `channel_ref`, `channel_epoch`,
`recipient_installation_ref`, `enc`, and `ciphertext/tag`. Detailed
provenance, key-storage observations, and the static iOS audit are in
`native/RESULTS.md`. Prior AES and
`BLOCKED-MTLS-WINDOWS-ENVIRONMENT` evidence is unchanged.

### BouncyCastle.Cryptography 2.6.2

- Source: official NuGet package and `bcgit/bc-csharp` source mirror.
- License: Bouncy Castle/MIT-style license as documented by the project.
- Maintenance: current stable package identified as 2.6.2; official C# source
  is maintained and targets modern .NET-compatible frameworks.
- C# API audit: no RFC 9180 HPKE Base-mode API exposing the exact frozen suite,
  caller-supplied info/AAD, and raw X25519 SerializePublicKey-compatible
  operations was found. HPKE support visible in Bouncy Castle's Java material
  was not treated as C# support.
- Decision: `REJECTED`; not added as a dependency and not emulated.

No second candidate met the required discovery threshold of public source,
explicit RFC 9180 claim, acceptable license, maintenance evidence, exact
suite, .NET 10 usability, and credible `net10.0-ios` path. No package audit
was run because no third-party package was used.

## AES-256-GCM project fixture

Independent Go and Windows .NET implementations produced byte-identical
output for the exact fixture:

- nonce Base64: `oKGio6Slpqeoqaqr`
- ciphertext Base64: `nToaRD2/d80HR73xZBWuqBXVOH7x0m8f+W1T9BbfDCy7GDOa3U0jECmtJrU=`
- tag Base64: `rG9A0yq3AsIzAvXMeaJyuw==`
- exact canonical AAD and plaintext are emitted by the harness and committed
  in `testdata/aes-gcm-project-fixture.json`.

Canonical fixture verification: `PASS`; ordinary runs never rewrite the
committed fixture. The Windows client independently uses
`System.Security.Cryptography.AesGcm`; the Go side uses `crypto/aes` and
`cipher.NewGCM`. Go opened the Windows-produced fixture and Windows opened the
Go-produced fixture: `PASS`.

AES stage: `AES-WINDOWS-INTEROP-PASS`.

- Go canonical-AAD tamper suite: `10/10 PASS`.
- Windows canonical-AAD tamper suite: `10/10 PASS`.
- Mutations covered: Trust Domain, Channel, format version, epoch, revision,
  Envelope reference, previous Envelope reference, nonce, ciphertext, and
  tag. Every case failed authentication without accepted plaintext or partial
  plaintext.

## Cleanup and validation

- Generated credentials/server keys are run-scoped and ignored; cleanup is in
  the orchestrator `finally` block.
- Full `run-windows.ps1`: `PASS` as a reproducible run, with stage
  classification `BLOCKED-HPKE`; its temporary directory was removed.
- `gofmt` verification: `PASS`.
- Go spike `go test ./...`: `PASS`.
- Go spike `go build -buildvcs=false`: `PASS`.
- .NET spike `dotnet restore`: `PASS`.
- .NET spike `dotnet build`: `PASS`.
- Vulnerability/dependency audit: not applicable; no third-party .NET package
  was used.
- Root `go vet ./...`: `PASS`.
- Root `go test ./...`: `PASS`.
- Root normal `go build ./...`: known sandbox VCS-stamping failure only;
  `go build -buildvcs=false ./...`: `PASS`.
- Root `go mod verify`: `PASS`.
- Root `git diff --check`: `PASS`.
