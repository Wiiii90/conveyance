# AGENTS.md

## Canonical project name

The canonical and only project name is **Conveyance**.

Do not use “Relay Service” as the project name. Relay is a technical/domain role inside Conveyance.

## Source of truth

Repository documentation and accepted ADRs are the durable source of truth.

Before implementation:

1. read `README.md`;
2. read the relevant `docs/` files and ADRs;
3. inspect the current `dev` state of Wiiii Got This, Vocation, and Illumination where the change crosses boundaries;
4. inspect the relevant published contracts;
5. identify contradictions before editing code;
6. do not infer domain ownership from deployment topology or implementation convenience.

## Boundary rules

Conveyance owns generic delivery semantics only.

Never introduce:

- Vocation domain models or business rules;
- Illumination learning/review/scheduling/reconciliation semantics;
- WGT capability-resolution or presentation semantics;
- shared databases across bounded contexts;
- cross-context ORM/domain imports;
- shared business-logic libraries that bypass published contracts;
- a generic “business object” API.

Foreign business payloads remain opaque and end-to-end protected.

## Security rules

Do not invent cryptographic constructions.

Before production security implementation:

- use only reviewed standard primitives/protocols;
- complete the .NET/iOS interoperability spike defined by ADR-0007;
- keep installation authentication, Channel authorization, key-encryption credentials, Channel Keys, and Recovery Material separate;
- never make server control sufficient for enrollment/recovery;
- never log private keys, Channel Keys, Recovery Material, decrypted foreign payloads, or raw credentials.

If a security requirement is not covered by an accepted ADR, stop implementation and return to the control plane.

## V1 scope

V1 implements only the `Current Object` Channel delivery mode.

Do not add:

- ordered streams;
- delta synchronization;
- CRDTs;
- generic conflict resolution;
- command queues;
- WebSockets/presence;
- push notifications;
- public accounts;
- generic service registry;
- Kafka/RabbitMQ/Kubernetes;
- product-visible snapshot history.

Future Vocation/Illumination bidirectional synchronization must use domain-owned contracts and reconciliation rules.

## Implementation workflow

Codex/Luna receives only narrow tasks whose semantics are already accepted.

Before each implementation task:

1. resolve current `dev` HEAD;
2. read affected contracts/ADRs;
3. name the exact vertical slice;
4. name required tests;
5. identify cross-context boundaries;
6. stop rather than invent missing architecture.

After implementation:

1. inspect the diff;
2. run complete relevant checks;
3. verify no plaintext foreign payload handling was introduced;
4. verify no foreign domain dependency was introduced;
5. verify scope did not expand;
6. update documentation only when the accepted design actually changed.
