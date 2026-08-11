---
title: startable and stoppable ordering
state: horizon
created: 2025-08-20
tags: [epic, spike]
subsystems: [da]
milestone: v1.0.x
source: github:michaelquigley/df#18
---

figure out an approach for controlling `Startable` and `Stoppable` ordering. the concrete container already reads `da:"order=N"` for field traversal; decide whether lifecycle ordering rides on that same tag, on a separate declaration, or on something derived from the wiring graph.
