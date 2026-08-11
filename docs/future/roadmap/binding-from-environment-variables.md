---
title: binding from environment variables
state: horizon
created: 2025-09-15
tags: [feature, spike]
subsystems: [dd]
milestone: v1.0.x
source: github:michaelquigley/df#30
---

implement a facility supporting generalized `Bind`/`Merge` from environment variables. `Merge` is the important half — it is what lets an application build a proper configuration cascade with the environment as the last layer over file-sourced defaults.
