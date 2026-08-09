# ADR-0002 -- Opaque End-to-End-Protected Foreign Payloads

Status: Accepted

## Decision

Foreign business payloads are protected client-side and remain opaque to
Conveyance.

Conveyance may see only minimal generic routing/delivery metadata
required for operation.

## Consequences

The server cannot validate Vocation or Illumination schemas and must not
log plaintext. TLS is necessary but insufficient; payload protection
remains end-to-end with respect to Conveyance.
