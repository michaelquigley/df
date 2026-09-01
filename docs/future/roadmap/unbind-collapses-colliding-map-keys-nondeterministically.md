---
title: unbind collapses colliding map keys nondeterministically
state: building
created: 2026-09-01
tags: [defect]
subsystems: [dd]
milestone: v1.0.x
log:
  - stamp: 2026-09-01
    note: vetoed in terminus review 25c161a328ba on the JSONL change — see docs/journal/2026-09-01.md
  - stamp: 2026-09-01
    note: terminus-canon `projects/df/dd-deterministic-serialization` now names this collision and carries a boundary exemption pointing at this card — remove that exemption paragraph when this lands
---

`valueToInterface`'s `reflect.Map` case (`dd/unbind.go`) stringifies every key through `keyToString` and assigns `result[keyStr]` without checking whether that spelling is already taken. interface-keyed maps can collide — `map[any]string{1: "int", "1": "str"}` yields two keys that both stringify to `"1"` via the `fmt.Sprintf("%v")` default — and whichever key go's randomized iteration visits last wins. `UnbindJSON`, `UnbindJSONL`, and `UnbindYAML` all inherit it, which breaks the deterministic-output guarantee documented in v1.0.2 for exactly this input.

detect the collision before assigning and return an error naming the map and the colliding spelling; a map that two keys serialize into cannot be represented losslessly, so failing is the honest answer over any tie-break. add a regression test on `map[any]V`. typed maps (`map[int]V`, `map[string]V`, …) cannot collide and need no change.

## why

this changes core `Unbind` behavior — a silent last-wins becomes an error — so it is compatibility work with its own release consideration, not a rider on a feature. vetoed on the JSONL change for the same reason the strict-mode findings of 2026-07-10 and 2026-07-13 were: the JSONL helpers neither introduce nor widen it. the reach is narrow (interface-keyed maps only), but the guarantee is written down and this is the one input that falsifies it.

## discussion

a 200-iteration probe of `UnbindJSONL` over `map[any]string{1: "int", "1": "str"}` split 176 × `{"m":{"1":"str"}}` against 24 × `{"m":{"1":"int"}}`.
