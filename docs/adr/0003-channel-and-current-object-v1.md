# ADR-0003 -- Opaque Channel with Current Object as the Only V1 Delivery Mode

Status: Accepted

## Decision

`Channel` is the generic opaque delivery/protection namespace.

V1 implements only `Current Object`: one logical current Envelope per
Channel, conditional atomic replacement, no product-visible history.

## Rationale

This exactly serves the first Vocation PC-off read flow without defining
Conveyance forever as snapshot-only infrastructure.

Future ordered/change delivery requires a new justified decision.
