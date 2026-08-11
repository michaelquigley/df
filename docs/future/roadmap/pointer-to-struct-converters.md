---
title: pointer-to-struct converters
state: horizon
created: 2026-07-13
tags: [defect, spike]
subsystems: [dd]
milestone: v0.1.x
source: docs/future/pointer-to-struct-converters.md
log:
  - stamp: 2026-07-13
    note: vetoed in terminus review 8ba11c9045b6 — see docs/journal/2026-07-13.md
---

the core binding layer does not consult a registered `Converter` before binding a pointer-to-struct field. a converter registered for `Temperature` that accepts `"21.5C"` runs when binding `Temperature`, but binding `*Temperature` classifies the target as a struct pointer and requires an object, so the scalar is rejected before the converter can run.

do not patch this for strict mode alone — that would make strict binding more capable than forgiving binding. census the equivalent direct, slice, and map binding paths for struct values and pointers, then pin the intended converter precedence in both modes with regression tests.

## why

surfaced during the strict-mode arc and vetoed there because it predates strict mode: this is a core binding limitation, not a strict-mode regression. a complete fix intentionally changes existing forgiving behavior, which makes it compatibility work with its own release consideration rather than a bug fix that rides along.
