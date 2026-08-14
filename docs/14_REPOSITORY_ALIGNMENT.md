# 14 – Repository Alignment Review

Reviewed against current `dev` state on 2026-08-14.

## Wiiii Got This

Verified architectural facts:

- Synchronization/Relay is already accepted as a separate target bounded context/service.
- personal Device trust/pairing without a mandatory account is accepted;
- Device identity, installation identity, cryptographic credentials, enrollment and revocation are conceptually distinct;
- hybrid recovery is accepted;
- server control alone must not become trust authority;
- WGT Personal Device Trust and Hybrid Recovery semantics are accepted;
- Conveyance's v0.2 Security Interoperability Profile is concrete and frozen for the
  interoperability proof;
- Go/Windows evidence is complete, while the physical iPhone proof remains open;
- Issue #6 remains the readiness gate and ADR-0007 remains the pre-production gate;
- production enrollment/revocation/recovery, payload integration, and other production-only
  security details remain gated; no Production Security approval is claimed.

Conveyance alignment:

- Conveyance consumes opaque Installation/Trust references and public credential state;
- it does not import WGT `DeviceIdentity` or `ServiceIntegration`;
- Channel authorization is a generic delivery ACL, not WGT Capability Resolution.

Current WGT alignment:

- WGT identifies Conveyance as the accepted owner of generic durable opaque delivery and
  retains device/platform integration and presentation ownership;
- Vocation `Published Opportunity Overview 1.0` is implemented and consumed by WGT Windows;
- WGT's remaining provider/runtime readiness gates are separate from Conveyance ownership.

No active WGT documentation drift relevant to Conveyance remains from the reviewed findings.

## Vocation

Verified:

- `Published Opportunity Overview 1.0` is implemented on `dev`;
- canonical schema: `schemas/published-opportunity-overview-v1.schema.json`;
- local read-only endpoint: `/published/v1/opportunity-overview`;
- endpoint remains outside the internal React OpenAPI;
- no relay, authentication, remote persistence, WGT client, or cross-device writes are implemented;
- Vocation remains locally authoritative and independently runnable.

Conveyance does not expose or interpret Vocation `publication_ref`, Opportunities, Companies, Postings, personal state, Freshness, or Availability.

ADR-0010 now preserves Vocation's local authority while describing Conveyance as the
separately accepted generic Synchronization/Relay bounded context. No current Vocation
alignment finding remains.

## Illumination

Verified:

- Illumination is local-first and locally authoritative;
- WGT is the primary Windows/iPhone presentation;
- future iPhone use with the PC off requires a device-local copy of learning data needed for study;
- remote readable learning persistence is not assumed;
- Illumination owns future domain-specific publication, change, command, authority, merge,
  conflict, scheduling and reconciliation semantics;
- WGT owns device/platform integration and presentation;
- Conveyance owns generic durable opaque cross-device delivery.

Conveyance alignment:

The generic Channel and per-Channel key/epoch model is compatible with future Illumination synchronization without defining its payload/change contract today.

`Current Object` remains the only accepted Conveyance V1 delivery mode; it is not an
automatically accepted bidirectional Learning synchronization solution. Illumination may
later justify an ordered/change delivery mode only through a separate domain contract and
system architecture decision.

## Review result

No cross-context ownership conflict blocks Conveyance.

The only implementation-blocking uncertainty found is the physical iPhone portion of concrete security interoperability. Go and Windows evidence is complete; Issue #6 remains open for the real .NET/iOS target and its Keychain, mTLS, HPKE, AES-GCM, and tamper cases. ADR-0007 remains an explicit pre-production gate: no partial result is production-ready security.
