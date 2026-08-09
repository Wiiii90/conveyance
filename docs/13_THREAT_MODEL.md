# 13 -- Threat Model

## Protected assets

-   foreign business payload confidentiality;
-   foreign payload integrity/authenticity;
-   Channel Keys and installation private keys;
-   recovery secret/material;
-   trusted-installation membership;
-   current Channel ordering/checkpoints.

## Server compromise

Assume attacker can read/modify Conveyance runtime, DB and backups.

Requirements: - no readable foreign payload; - no plaintext Channel
Keys/recovery secret; - server cannot enroll a new trusted installation
by itself; - payload/metadata tampering is detected client-side.

Availability cannot be guaranteed against a compromised server.

## Stolen database

Database/backups expose ciphertext, opaque identifiers, sizes, timing
and necessary routing metadata, but no business plaintext or client
private secrets.

## Network attacker

TLS 1.3 protects transport and authenticates server/client according to
the selected deployment. E2E payload protection remains necessary
because the server itself is outside the payload-confidentiality trust
boundary.

## Lost/stolen installation

Installation can be revoked. Future server access is rejected. Affected
Channels rekey to new epochs. Previously learned old keys/plaintext
cannot be revoked.

## Malicious new installation

Knowledge of server address, Trust Domain reference or network proximity
is insufficient. Enrollment requires an existing trusted authority or
Recovery Authority.

## Replay / rollback

AEAD authenticity alone does not make old valid Envelopes invalid.
Clients retain trusted Channel checkpoints and reject lower
epoch/revision state.

A brand-new/recovered client requires checkpoint state transferred
through trusted enrollment/recovery.

## Concurrent overwrite

Current Object replacement uses compare-and-swap against expected
current state. A stale concurrent writer receives Conflict.

## Metadata exposure

V1 accepts exposure of minimal generic metadata: - opaque Trust
Domain/Channel/Envelope refs; - epoch/revision; - sizes; - server
timing; - authenticated installation identity during requests.

Foreign service/capability/business identifiers are not required
visible.

Traffic analysis and complete metadata hiding are not V1 goals.

## Denial of service

Server can refuse/delete data or be offline. Mitigation is local-first
operation and validated local caches, not Byzantine availability
infrastructure.

## Server rollback

Established clients detect rollback below their trusted checkpoint.
After total loss of all independent checkpoints, an authentic old
Recovery Package may be difficult to distinguish from current state. V1
accepts this residual risk rather than introducing transparency
infrastructure.
