# 09 -- Read Models

Conveyance V1 intentionally has few read models.

## Current Channel State

For an authorized caller: - channel reference; - current epoch; -
current revision; - current envelope reference; - current opaque
Envelope/protected payload where requested.

It does not expose foreign payload fields.

## Installation Access State

Administrative/security-facing projection: - installation reference; -
status; - generic channel permissions; - public credential metadata
necessary for operation.

No foreign business state.

## Operational state

Server operators may observe: - request counts; - response/error
categories; - payload byte sizes; - storage consumption; - timestamps; -
latency.

Logs/metrics must not contain decrypted foreign payloads, private keys,
Channel Keys, recovery secrets, or raw authentication secrets.
