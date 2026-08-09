# ADR-0001 -- Conveyance as Separate Synchronization/Relay Bounded Context

Status: Accepted

## Decision

Conveyance is a separate bounded context and independently deployable
microservice.

It owns generic delivery semantics only.

## Consequences

WGT, Vocation and Illumination remain independent bounded contexts. No
shared database, foreign domain imports, or shared business-logic
library is introduced. Conveyance failure does not invalidate local
domain operation.
