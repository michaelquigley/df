---
title: dynamic and strict do not compose
state: inbox
created: 2026-08-10
tags: [defect]
subsystems: [dd]
source: llm-gateway
---

the `type` discriminator that `dd` requires for a `Dynamic` field is then rejected by `dd`'s own strict binder as an unknown field:

```
Holder.Sources[0]: binding Dynamic type "file" failed: SrcCfg: unknown field "type"
```

have the strict binder treat the discriminator key as consumed for the struct a `DynamicBinder` produces, the same way `+extra` opts a struct out. `dd` put the key there; it should account for it.

## why

the `Strict()` doc comment says Dynamic binders "remain in effect under strict mode," so the current behavior reads as a bug rather than a boundary. workarounds exist — the binder can `delete(m, "type")`, or the target can carry a throwaway `Type` field — but both are the caller compensating for `dd` rejecting a key `dd` mandated. it will bite the first person decoding a discriminated union out of a signed payload.

## discussion

the repro:

```go
opts := dd.Strict()
opts.DynamicBinders = map[string]func(map[string]any) (dd.Dynamic, error){
    "file": func(m map[string]any) (dd.Dynamic, error) {
        v := &SrcCfg{}                      // Name, Path — no Type field
        if err := dd.Bind(v, m, dd.Strict()); err != nil {
            return nil, err
        }
        return v, nil
    },
}
dd.BindYAML(h, []byte("sources:\n  - type: file\n    name: local\n    path: /x\n"), opts)

→ Holder.Sources[0]: binding Dynamic type "file" failed: SrcCfg: unknown field "type"
```

not blocking llm-gateway, which surfaced it while adopting `dd.Strict()` on a `v1.0.1` → `v1.0.2` bump: their heterogeneous `sources` list is operator-authored configuration rather than an external contract, so it decodes under the forgiving posture and never hits this in anger.
