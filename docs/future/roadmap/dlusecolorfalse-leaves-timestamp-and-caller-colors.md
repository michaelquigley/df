---
title: DL_USE_COLOR=false leaves timestamp and caller colors
state: researching
created: 2026-09-23
tags: [defect]
subsystems: [dl]
milestone: v1.1.x
---

with `DL_USE_JSON=false DL_USE_COLOR=false`, the pretty handler still wraps the timestamp, the caller, the channel, and the fields in ANSI escapes; only the level labels honor `UseColor`. `DefaultOptions` wraps the labels conditionally, but `handler.go` writes `TimestampColor`, `FunctionColor`, `ChannelColor`, and `FieldsColor` unconditionally. make every color in the pretty format follow `UseColor`, so a daemon running under systemd with both variables set gets plain text in the journal.

## background

found on astrometrics 2026-09-23: the user units set both variables per the convention and journald still receives `\e[90m` and `\e[36m` around every line. reproduce with `DL_USE_JSON=false DL_USE_COLOR=false <any dl program> 2>&1 | cat -v`.
