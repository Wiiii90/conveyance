# 03 -- Ubiquitous Language

## Trust Domain

An opaque personal trust namespace. It is not a user account and
contains no required email/username identity.

## Installation

A cryptographically enrolled client installation that can authenticate
to Conveyance. It is not identical to WGT's current `DeviceIdentity` and
is not identical to a cryptographic key.

## Channel

An opaque, durable delivery and protection namespace inside a Trust
Domain.

A Channel carries no server-visible Vocation/Illumination meaning. Local
adapters map foreign capabilities or synchronization purposes to Channel
references.

## Delivery Mode

The generic delivery behavior of a Channel. V1 supports only
`Current Object`.

## Current Object

A delivery mode in which a Channel exposes exactly one logical current
Envelope. Replacing it does not create product-visible history.

## Envelope

One protected opaque payload plus Conveyance-owned delivery/protection
metadata.

## Envelope Reference

Opaque identity of an Envelope. It is not a Vocation publication
reference or foreign domain identifier.

## Channel Epoch

A monotonically advancing cryptographic generation of a Channel.
Rekey/revocation can advance the epoch.

## Revision

Monotonic Current Object revision within one Channel Epoch.

## Channel Key

A secret owned by trusted clients for one Channel Epoch. Conveyance
never receives it in plaintext.

## Key Grant

A protected representation that permits one authorized installation to
obtain a Channel Key.

## Installation Authentication Credential

Credential used to authenticate an installation to Conveyance. It is
separate from payload-decryption/key-encryption material.

## Recovery Authority / Recovery Material

Independent high-entropy secret material capable of recovering the Trust
Domain and required Channel-key state after loss of all trusted
installations.

## Checkpoint

The highest trusted `(channel_epoch, revision, envelope_ref)` state
known locally for a Channel.

## Publication / Business Payload

Foreign-context content protected inside an Envelope. Conveyance treats
it as bytes and does not interpret it.
