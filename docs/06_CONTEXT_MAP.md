# 06 -- Context Map

## Wiiii Got This → Conveyance

WGT owns integration, platform/device presentation, and personal
device-trust orchestration.

WGT adapters: - consume foreign published/application contracts; -
validate them; - protect/deprotect opaque payloads; - map local
integration purposes to opaque Conveyance Channels; - validate foreign
contracts again after decryption.

Conveyance owns generic delivery only.

No WGT domain classes are imported into Conveyance.

## Vocation → WGT → Conveyance

``` text
Vocation
  │ Published Opportunity Overview 1.0
  ▼
WGT Vocation Integration Adapter
  │ validate + protect
  ▼
Conveyance Channel (opaque)
  │ Current Object
  ▼
WGT target installation
  │ decrypt + validate
  ▼
Vocation Published Contract
```

Vocation remains standalone and locally authoritative. It does not
require a Conveyance client dependency for the first flow.

## Illumination → WGT/adapter → Conveyance

Illumination remains locally authoritative for learning content,
Reviews, Scheduling, Learning State and Progress.

Future synchronization:

``` text
Illumination-owned sync payload
  ▼
WGT / Illumination sync adapter
  ▼
opaque Conveyance Channel
  ▼
target installation
  ▼
Illumination-owned reconciliation
```

Conveyance never owns reconciliation.

## Separate Ways where appropriate

Absence of Conveyance does not prevent local operation of WGT
integrations, Vocation, or Illumination.
