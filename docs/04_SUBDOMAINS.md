# 04 -- Subdomains

Conveyance remains one bounded context and one deployable service.

Logical responsibilities are not separate microservices.

## Delivery

Core domain responsibility:

-   Channel;
-   Current Object;
-   Envelope;
-   revision/epoch ordering;
-   compare-and-swap replacement;
-   retention.

## Trust-facing access

Supporting responsibility:

-   authenticated Installation principal;
-   generic Channel grants (`read`, `publish`);
-   revocation enforcement.

Trust establishment itself is not server-owned authority: enrollment
approval and recovery authority remain client-controlled.

## Cryptographic transport support

Supporting responsibility:

-   opaque key-grant persistence;
-   envelope protection metadata;
-   protocol/version validation.

Conveyance does not perform business-payload decryption or server-side
re-encryption.

## Operations

Generic infrastructure responsibility:

-   storage;
-   quotas/size limits;
-   HTTP/API errors;
-   observability without business-payload logging.
