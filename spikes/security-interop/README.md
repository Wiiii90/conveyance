# Security interoperability spike

This directory contains empirical v0.2.0 spike code only. It is not imported
by the production Conveyance executable and does not define production APIs.

Run `run-windows.ps1` from this directory on Windows. The orchestrator uses
Go 1.26 and .NET 10, creates a run-scoped temporary directory, runs the Go
AES/HPKE checks and Go TLS server, exercises the Windows certificate store,
and removes generated credentials in a `finally` block.

The Windows HPKE stage is intentionally not emulated. If the audited managed
candidate is unavailable or lacks RFC 9180 support, `RESULTS.md` records
`BLOCKED-HPKE`; no custom X25519/HKDF/AES construction is introduced.
