# ADR-0006 -- Client Checkpoints for Replay/Rollback Detection

Status: Accepted

## Decision

Clients retain the highest trusted
`(channel_epoch, revision, envelope_ref)` per Channel and reject
regressions.

Pairing/recovery may transfer trusted checkpoints.

## Residual risk

After total loss of all independent trusted checkpoints, a fully
malicious server may replay an older but authentic recovery state. V1
accepts this limitation and does not introduce a transparency service.
