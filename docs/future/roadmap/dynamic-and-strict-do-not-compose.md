---
title: dynamic and strict do not compose
state: inbox
created: 2026-08-10
tags: [defect]
subsystems: [dd]
source: docs/future/requests.md
---

the `type` discriminator that `dd` requires for a `Dynamic` field is then rejected by `dd`'s own strict binder as an unknown field:

```
Holder.Sources[0]: binding Dynamic type "file" failed: SrcCfg: unknown field "type"
```

have the strict binder treat the discriminator key as consumed for the struct a `DynamicBinder` produces, the same way `+extra` opts a struct out. `dd` put the key there; it should account for it.

## why

the `Strict()` doc comment says Dynamic binders "remain in effect under strict mode," so the current behavior reads as a bug rather than a boundary. workarounds exist — the binder can `delete(m, "type")`, or the target can carry a throwaway `Type` field — but both are the caller compensating for `dd` rejecting a key `dd` mandated. it will bite the first person decoding a discriminated union out of a signed payload.
