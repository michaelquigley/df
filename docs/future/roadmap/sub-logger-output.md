---
title: sub-logger output
state: horizon
created: 2026-03-18
tags: [enhancement, spike]
subsystems: [dl]
milestone: v1.0.x
source: github:michaelquigley/df#49
---

let `dl` detect that it is running inside a larger logging system — a `DL_SUB_LOGGER` environment variable was the first sketch — and emit output that composes well with the host instead of defaulting to JSON. the open question is what "looks good and works well" means concretely for a nested logger, since the host's format is unknown to us.
