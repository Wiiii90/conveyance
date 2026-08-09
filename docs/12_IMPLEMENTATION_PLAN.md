# 12 -- Implementation Plan

## Control-plane rule

GitHub milestones are version named, and GitHub Issues are the durable work
packages within them. The Control Plane owns scope, sequencing, and
parallelization. Workers do not expand these packages while implementing
them.

## v0.1.0 -- Current Object baseline

v0.1.0 is sequenced through these work packages:

1. Issue #1, repository bootstrap: Go module, lifecycle, CI, and container
   baseline. Completed.
2. Issue #2, contract freeze: exact Current Object HTTP shapes, opaque UUID
   references, transitions, atomic CAS, payload/error semantics, and the
   temporary authorization seam. This documentation slice must complete
   before Issues #3--#5 implement the contract.
3. Issue #3, domain and application core: opaque value types, immutable
   Envelope, Current Object invariants and transitions, persistence-independent
   operations, and unit tests.
4. Issue #4, durable persistence: minimal implementation selection, atomic CAS,
   restart durability, no product-visible history, and persistence integration
   tests. Stop if technology selection requires an unaccepted architecture
   decision.

5. Issue #5, HTTP API: the two versioned Current Object endpoints, frozen JSON
   and generic error mapping, decoded payload-size enforcement,
   application/persistence integration, local/test authorization adapter, and
   HTTP/integration tests.

Issue #4 selects SQLite through `database/sql` with `modernc.org/sqlite` for
the v0.1.0 server. It is a small single-node durable Current Object store, and
the driver avoids CGO. Persistence is Conveyance-owned only; there is no shared
database, ORM, or migration framework in v0.1.0.

v0.1.0 implements no production authentication or cryptography. It has no
separate Channel-create endpoint: first publish atomically establishes the
`current_object` Channel.

## Issue #2 gate

Issues #3--#5 must implement, not redesign, the contract frozen in ADR-0003 and
`docs/08_CONTRACTS.md`. Any contradiction or missing architecture decision
returns to the Control Plane before code changes.

## Later security integration

A future Control-Plane-assigned milestone may implement accepted TLS/client
authentication integration, installation credential lifecycle, Channel
grants, opaque Channel Key Grants, revocation enforcement, recovery-package
persistence, and client interoperability.

Before any production security claim, execute ADR-0007 as a narrow technical
spike on the actual .NET Windows/iPhone and Go runtimes. If HPKE or mTLS cannot
be made reliable, stop and return to the Control Plane. Do not invent a custom
cryptographic or authentication protocol.

## Later foreign-context vertical proof

Vocation/WGT/Illumination integration remains outside v0.1.0. Any later
vertical slice uses domain-owned published contracts and client-side
protection; Conveyance stores and returns opaque bytes only.

## Explicitly deferred

- bidirectional Vocation writes;
- Illumination synchronization;
- ordered/change delivery mode;
- acknowledgement/retry product semantics;
- delta synchronization;
- CRDTs;
- generic conflict resolution;
- push/presence;
- web UI;
- public accounts;
- service registry;
- Kafka, RabbitMQ, or Kubernetes.
