---
title: unbind serializes identity-bearing map keys as runtime addresses
state: inbox
created: 2026-09-01
tags: [defect]
subsystems: [dd]
log:
  - stamp: 2026-09-01
    note: raised blocking in terminus review c3aea7e682c1 on the key-collision fix, vetoed there as pre-existing — see docs/journal/2026-09-01.md
---

`keyToString` (`dd/convert.go`) formats every map key kind outside string, int, uint, float, and bool through `fmt.Sprintf("%v")`. pointer, channel, func, and unsafe-pointer keys — reached directly, through an interface key, or inside a struct or array key — therefore serialize as runtime addresses (`0xc000012340`), which differ across processes for the same logical input. the collision check from `unbind-collapses-colliding-map-keys-nondeterministically` cannot catch this: distinct addresses have distinct spellings, so the output is well-formed and still falsifies the cross-process byte guarantee for `UnbindJSON`, `UnbindJSONL`, and `UnbindYAML`.

reject identity-bearing key kinds at unbind with an `UnsupportedError` before stringification, composites included. `stringToKey` on the bind side accepts only the scalar kinds, so such maps were never round-trippable; refusing them aligns unbind with what bind can produce. add a regression test on `map[*int]V`.

## why

pre-existing and outside the collision card, so it was vetoed there for the same reason that card was vetoed on the JSONL change: the fix neither introduced nor widened it. it is a second behavior change to forgiving `Unbind` — address-keyed maps that serialized now error — and carries its own release consideration.
