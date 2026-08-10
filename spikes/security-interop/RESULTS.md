# v0.2.0 Go + Windows spike results

Run date: 2026-08-10

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
- Selected Go 1.26.5 standard-library vector material: recipient key and
  encapsulation material parse/validate, `PASS`.
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
- Result: `FAIL` in this sandbox before credential creation. The Windows CNG
  call fails at `CngKey.Create` with
  `CryptographicException: The system cannot find the file specified.`
- Therefore persistence across process boundaries, TLS 1.3 negotiation,
  registered success, unknown rejection, and InstallationRef mismatch could
  not be executed here. The client code does not fall back to an exportable or
  in-memory credential.
- Server implementation is loopback-only, TLS 1.3-only, requires a client
  certificate, pins the test server SPKI in the client, and checks the explicit
  InstallationRef mapping. No trust-all callback or production CA is used.

## .NET HPKE candidate audit

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

Fixture equality: `PASS`. The Windows client independently uses
`System.Security.Cryptography.AesGcm`; the Go side uses `crypto/aes` and
`cipher.NewGCM`. Go opened the Windows-produced fixture and Windows opened the
Go-produced fixture: `PASS`. The complete Windows tamper suite remains
pending the mTLS/HPKE executable client path.

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
- Production repository checks and final `git diff --check` remain to be run
  after the spike artifacts are complete.
