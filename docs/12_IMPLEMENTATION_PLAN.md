# 12 -- Implementation Plan

## Gate

No production implementation begins until: - this specification set is
reviewed; - security ADRs are accepted; - concrete crypto/library
interoperability is verified; - WGT/Vocation documentation drift is
synchronized or explicitly tracked.

## Milestone 0 -- Repository bootstrap

After architecture approval: - create `Wiiii90/conveyance`; - `main`
stable branch and `dev` active development branch; - add specification
documents and ADRs; - Go module/bootstrap; - lint/test/CI baseline; -
minimal container build.

No business-payload semantics.

## Milestone 1 -- Current Object relay core

Implement: - Trust Domain/Installation server representations required
by protocol; - opaque Channel with `current_object`; - Envelope
validation; - atomic CAS replacement; - authorized current retrieval; -
durable storage; - size limits and generic errors; - tests.

## Milestone 2 -- Security integration

Implement accepted: - TLS/client authentication integration; -
installation credential lifecycle; - Channel grants; - opaque HPKE-style
Channel Key Grants; - revocation enforcement; - recovery-package
persistence.

Client-side crypto belongs in appropriate WGT/security
adapters/libraries, not as server decryption logic.

## Milestone 3 -- Vocation vertical proof

Cross-repo slice: - WGT Windows consumes/validates Vocation Published
Opportunity Overview 1.0; - protects it into a Conveyance Envelope; -
publishes Current Object; - WGT iPhone retrieves, verifies/decrypts,
validates Vocation contract, caches; - prove Windows/Vocation can be
offline during iPhone read.

## Explicitly deferred

-   bidirectional Vocation writes;
-   Illumination synchronization;
-   ordered/change delivery mode;
-   ack/retry product semantics;
-   delta sync;
-   CRDTs;
-   generic conflict resolution;
-   push/presence;
-   web UI;
-   public accounts;
-   service registry;
-   Kafka/RabbitMQ/Kubernetes.


## Pre-production interoperability spike

Before Milestone 2 can claim production security, execute ADR-0007 as a narrow technical spike.

This spike is allowed to use isolated test code and fixtures. It must not expand into the production synchronization implementation.

If HPKE or mTLS cannot be made reliable on the actual .NET iPhone target, stop and return to the control plane with measured results.
