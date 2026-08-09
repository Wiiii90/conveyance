# ADR-0004 -- Per-Channel Key Hierarchy with Epoch Rekey

Status: Accepted

## Decision

Each Channel Epoch has its own Channel Key. Channel Keys are protected
individually for authorized installations and Recovery Authority. A
revoked installation receives no new-epoch key material.

No global Trust-Domain business master key is used to encrypt all
service data.

## Consequences

Blast radius is limited by Channel. Revocation rotates active
synchronized Channels to fresh epochs. Old data already accessible to a
revoked device cannot be made secret retroactively.
