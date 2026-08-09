# 01 -- Domain Vision

## Purpose

Conveyance provides generic durable delivery between trusted
installations in the We Got This! system when producer and consumer
devices are not simultaneously online.

Its primary architectural purpose is cross-device continuity without
making the personal server an authority over foreign domain data or
personal device trust.

## Bounded-context responsibility

Conveyance owns:

-   opaque Channel identity;
-   Envelope identity;
-   generic delivery modes;
-   Current Object semantics in V1;
-   generic sequencing and optimistic concurrency;
-   durable encrypted-envelope storage;
-   transport authentication and generic authorization;
-   retention;
-   generic delivery errors and technical limits;
-   generic key-grant storage required by the security protocol.

Conveyance does not own:

-   Vocation Opportunities, Companies, Postings, Assessments, Decisions,
    Availability/Freshness, or reconciliation;
-   Illumination Learning Items, Reviews, Scheduling, Learning State,
    Progress, or reconciliation;
-   WGT capability resolution, device-specific presentation, or
    service-integration domain semantics.

## Strategic constraints

-   Foreign business payloads are opaque to Conveyance.
-   Remote readable business persistence is not required.
-   Local services remain usable when Conveyance is absent or
    unavailable.
-   The first slice must remain small.
-   No speculative event bus, CRDT platform, command queue, presence
    service, or generic backup product is introduced.
