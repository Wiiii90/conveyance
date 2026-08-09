# 14 – Repository Alignment Review

Reviewed against current `dev` state on 2026-08-09.

## Wiiii Got This

Verified architectural facts:

- Synchronization/Relay is already accepted as a separate target bounded context/service.
- personal Device trust/pairing without a mandatory account is accepted;
- Device identity, installation identity, cryptographic credentials, enrollment and revocation are conceptually distinct;
- hybrid recovery is accepted;
- server control alone must not become trust authority;
- exact crypto/key hierarchy was intentionally deferred to the first concrete synchronized flow.

Conveyance alignment:

- Conveyance consumes opaque Installation/Trust references and public credential state;
- it does not import WGT `DeviceIdentity` or `ServiceIntegration`;
- Channel authorization is a generic delivery ACL, not WGT Capability Resolution.

Known WGT documentation drift to synchronize separately:

- `docs/23_FOREIGN_CONTEXT_ALIGNMENT.md` still describes the old Vocation mobile read-model direction and says no production WGT read contract exists;
- `docs/22_DEFERRED_DECISIONS.md` still treats the first Vocation WGT contract as future work;
- `docs/10_ARCHITECTURE.md` contains stale Vocation-readiness text and older “still required” architecture decisions;
- README contains an old specification-only status section above the accepted/implemented bootstrap baseline;
- ADR-0002 still predicts Illumination as the likely first concrete synchronized flow.

These are documentation synchronization items, not Conveyance domain changes.

## Vocation

Verified:

- `Published Opportunity Overview 1.0` is implemented on `dev`;
- canonical schema: `schemas/published-opportunity-overview-v1.schema.json`;
- local read-only endpoint: `/published/v1/opportunity-overview`;
- endpoint remains outside the internal React OpenAPI;
- no relay, authentication, remote persistence, WGT client, or cross-device writes are implemented;
- Vocation remains locally authoritative and independently runnable.

Conveyance does not expose or interpret Vocation `publication_ref`, Opportunities, Companies, Postings, personal state, Freshness, or Availability.

Known Vocation documentation drift:

ADR-0010 says Relay/Storage “is not a Sync bounded context” and that no separate Sync bounded context is introduced. This must be clarified to mean only that **Vocation does not introduce or own one**. System-wide, Conveyance is the separately accepted Synchronization/Relay bounded context.

## Illumination

Verified:

- Illumination is local-first and locally authoritative;
- WGT is the primary Windows/iPhone presentation;
- future iPhone use with the PC off requires a device-local copy of learning data needed for study;
- remote readable learning persistence is not assumed;
- Illumination owns future domain-specific change, merge, conflict, scheduling and reconciliation semantics;
- generic infrastructure may own relay/transport/retry/encryption.

Conveyance alignment:

The generic Channel and per-Channel key/epoch model is compatible with future Illumination synchronization without defining its payload/change contract today.

`Current Object` remains the only Conveyance V1 delivery mode. Illumination may later justify an ordered/change delivery mode through a separate contract decision.

## Review result

No cross-context ownership conflict blocks Conveyance.

The only implementation-blocking uncertainty found is concrete security interoperability across Go and WGT .NET/iOS. ADR-0007 turns that uncertainty into an explicit pre-production gate rather than an implicit assumption.
