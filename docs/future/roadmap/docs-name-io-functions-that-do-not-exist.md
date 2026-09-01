---
title: docs name io functions that do not exist
state: inbox
created: 2026-09-01
tags: [documentation]
subsystems: [dd]
log:
  - stamp: 2026-09-01
    note: every in-repo instance (dd/README.md, the docs-site dd and df-framework guides, the dd_04_io example README) was fixed on terminus findings across reviews 7659d6ec4a91, 937c6f0da002, 425465e46df3 — see docs/journal/2026-09-01.md; only the out-of-repo AGENTS.md source remains
---

`AGENTS.md` — a symlink into `tools/agents/df.md`, outside this repo — documents file i/o as `BindFromJSON/YAML()` and `UnbindToJSON/YAML()`. neither exists; they were renamed after the release that `CHANGELOG.md` records as introducing `NewFromYAML`/`MergeFromYAML`. the real surface is `dd/io.go`: `BindJSONFile`/`BindYAMLFile`, `NewJSONFile[T]`/`NewYAMLFile[T]`, `UnbindJSONFile`/`UnbindYAMLFile`, plus `UnbindJSONL`/`UnbindJSONLWriter` for JSON Lines. fix the bullet at its source in `tools/agents/df.md`.

## why

every agent boots from `AGENTS.md`; stale names there mean each one arrives with a wrong map of the i/o surface and has to rediscover the real one from `io.go`.
