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
