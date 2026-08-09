# 11 -- Acceptance Tests

## Boundary tests

1.  Server persistence contains no plaintext Vocation Published
    Opportunity Overview.
2.  Conveyance code has no dependency on Vocation/Illumination/WGT
    domain models.
3.  Unknown foreign payload bytes can be stored/retrieved without
    interpretation.

## Current Object tests

4.  First authorized publish creates current Envelope.
5.  Authorized read returns exactly that Envelope.
6.  Successful conditional replacement makes the new Envelope current.
7.  Competing replacement based on stale expected state returns
    Conflict.
8.  No product API exposes superseded history in V1.

## Authorization tests

9.  Unauthenticated request is rejected.
10. Revoked installation authentication is rejected.
11. Active installation without `read` cannot retrieve.
12. Active installation without `publish` cannot replace.
13. Read-only iPhone can retrieve Vocation V1 Channel.

## Cryptographic protocol tests

14. Ciphertext modification is detected by the client.
15. Changing authenticated Channel/epoch/revision/envelope metadata
    causes verification failure.
16. Conveyance cannot derive Channel Key from stored Key Grants.
17. New epoch uses a fresh Channel Key.
18. Revoked installation receives no new-epoch Key Grant.

## Rollback tests

19. Client that accepted revision N rejects lower revision in same
    epoch.
20. Client that accepted epoch E rejects lower epoch.
21. Pairing can transfer a trusted current checkpoint to a new
    installation.

## Recovery tests

22. Server alone cannot recover Channel Keys.
23. Correct Recovery Material can recover the current protected recovery
    state.
24. Incorrect Recovery Material cannot.
25. Recovery can restore access to latest recoverable Current Object
    after total device loss.
26. Recovery rollback limitation after loss of all external checkpoints
    is explicitly tested/documented where simulation permits.

## Availability tests

27. Conveyance outage does not make local Vocation unusable.
28. Previously validated local WGT cache remains usable according to WGT
    policy.
