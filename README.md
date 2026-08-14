# Conveyance

Conveyance is the synchronization/relay bounded context of the We Got
This! system.

It provides generic, durable, end-to-end-protected delivery between
trusted installations without interpreting foreign business payloads.

Conveyance is independent from Wiiii Got This (WGT), Vocation, and
Illumination. It owns delivery semantics only. It never owns foreign
domain semantics, reconciliation rules, or presentation.

## V1 goal

The first vertical proof is the cross-device read of Vocation's
`Published Opportunity Overview 1.0` while the publishing Windows PC is
offline:

1.  Vocation publishes its local read-only contract.
2.  WGT Windows validates it.
3.  WGT protects the complete payload client-side.
4.  WGT publishes an opaque envelope to Conveyance.
5.  Conveyance durably stores the current envelope.
6.  WGT iPhone authenticates and retrieves it.
7.  WGT iPhone verifies/decrypts it locally.
8.  WGT validates the original Vocation contract and caches it locally.

The server cannot read the Vocation payload.

## Current implementation status

The v0.1.0 local/test baseline implements:

- the Go Current Object service;
- opaque Channel/Envelope transport semantics;
- SQLite durable current-state persistence;
- the GET/PUT Current Object HTTP API;
- atomic compare-and-swap (CAS);
- an 8 MiB decoded payload default.

This baseline is intentionally local/test-only. The runtime binds to
`127.0.0.1:8080` and uses the runtime database `conveyance.db`. Its current
allow-all local/test `OperationContext` is not production authentication.
The Go and Windows evidence for the frozen ADR-0007 security-interoperability
path is complete. The remaining gate is the physical real-iPhone proof:
Keychain-backed installation authentication, TLS 1.3 mTLS rejection cases,
HPKE Go↔iPhone interoperability and tamper rejection, and the frozen
AES-256-GCM Envelope fixture and tamper cases. Issue #6 remains open; a
simulator or Windows iOS build is not a substitute. Production mTLS,
enrollment, revocation, recovery, and payload-protection integration remain
deferred, and no production-security claim is made. Conveyance does not
itself prove that arbitrary uploaded payload bytes are encrypted; client-side
protection belongs to the accepted security integration.

## Direction

Preferred implementation stack: Go.

V1 implements only the `Current Object` delivery mode. The model
deliberately permits future generic delivery modes without implementing
them now.

Core principles:

-   opaque end-to-end-protected payloads;
-   explicit versioned contracts;
-   local-first operation;
-   independent bounded contexts and releases;
-   no shared database or foreign domain imports;
-   no domain-specific merge/conflict semantics in Conveyance;
-   server control alone cannot establish personal device trust;
-   revocation and hybrid recovery are first-class security
    requirements.
